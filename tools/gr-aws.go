package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fatih/color"
	tld "github.com/jpillora/go-tld"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/D3Ext/go-recon/core"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var cyan func(a ...interface{}) string = color.New(color.FgCyan).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()
var yellow func(a ...interface{}) string = color.New(color.FgYellow).SprintFunc()

const s3_url = ".s3.amazonaws.com"

var found_buckets []string

type BucketsInfo struct {
	Buckets []string `json:"buckets"`
	Length  int      `json:"length"`
	Time    string   `json:"time"`
}

func helpPanel() {
	fmt.Println(`Usage of gr-aws:
    -d)       target domain to find its S3 AWS buckets and their permissions (i.e. example.com)
    -l)       file containing a list of domains to find their S3 AWS buckets and their permissions (one domain per line)
    -b)       file containing a list of bucket names to check if they exist, and to list permissions (one name per line)
    -p)       file containing a list of permutations, if not especified, default perms are used (doesn't apply with -b)
    -w)       number of concurrent workers (default=5)
    -o)       file to write buckets urls into
    -oj)      file to write buckets urls into (JSON format)
    -c)       print colors on output (recommended)
    -q)       don't print banner nor extra logging, only output
    -h)       print help panel
  
Examples:
    gr-aws -d example.com -o buckets.txt -c
    gr-aws -l domains.txt -p perms.txt -w 15
    cat domains.txt | gr-aws
    `)
}

func main() {
	var domain string
	var list string
	var buckets_file string
	var perms_list string
	var workers int
	var output string
	var json_output string
	var quiet bool
	var stdin bool
	var use_color bool
	var help bool

	flag.StringVar(&domain, "d", "", "target domain to find its S3 AWS buckets and their permissions (i.e. example.com)")
	flag.StringVar(&list, "l", "", "file containing a list of domains to find their S3 AWS buckets and their permissions (one domain per line)")
	flag.StringVar(&buckets_file, "b", "", "file containing a list of bucket names to check if they exist, and to list permissions (one name per line)")
	flag.StringVar(&perms_list, "p", "", "file containing a list of permutations (if not especified, default perms are used)")
	flag.IntVar(&workers, "w", 5, "number of concurrent workers")
	flag.StringVar(&output, "o", "", "file to write buckets urls into")
	flag.StringVar(&json_output, "oj", "", "file to write buckets urls into (JSON format)")
	flag.BoolVar(&quiet, "q", false, "don't print banner, only output")
	flag.BoolVar(&use_color, "c", false, "print colors on output")
	flag.BoolVar(&help, "h", false, "print help panel")
	flag.Parse()

	t1 := core.StartTimer()

	if !quiet {
		fmt.Println(core.Banner())
	}

	if help {
		helpPanel()
		os.Exit(0)
	}

	// Check if stdin has value
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		stdin = false
	} else {
		stdin = true
	}

	// if domain, list and stdin parameters are empty print help panel and exit
	if (domain == "") && (list == "") && (buckets_file == "") && (!stdin) {
		helpPanel()
		core.Red("Especify a valid argument (-d), (-l) or (-b)", use_color)
		os.Exit(0)
	}

	if ((domain != "") && (list != "")) || ((domain != "") && (buckets_file != "")) || ((list != "") && (buckets_file != "")) {
		helpPanel()
		core.Red("You can't use (-d), (-l) or (-b) at same time", use_color)
		os.Exit(0)
	}

	// define variables which will be used to write buckets to output files
	var out_f *os.File
	if output != "" {
		out_f, err = os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
	} else if json_output != "" {
		out_f, err = os.Create(json_output)
		if err != nil {
			log.Fatal(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if json_output != "" {
			json_buckets := BucketsInfo{
				Buckets: found_buckets,
				Length:  len(found_buckets),
				Time:    core.TimerDiff(t1).String(),
			}

			json_body, err := json.Marshal(json_buckets)
			if err != nil {
				log.Fatal(err)
			}

			_, err = out_f.WriteString(string(json_body))
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(0)
	}()

	var perms []string
	if perms_list != "" {
		pf, err := os.Open(perms_list) // Get payloads from given file
		if err != nil {
			log.Fatal(err)
		}
		defer pf.Close()

		sc := bufio.NewScanner(pf)
		for sc.Scan() {
			perms = append(perms, sc.Text())
		}

	} else {
		perms = core.GetPerms() // Default permutations list (285 words)
	}

	var counter int
	var wg sync.WaitGroup
	buckets_c := make(chan string) // Create channel and wait group for concurrency

	if !quiet {
		core.Magenta("Looking for S3 buckets with "+strconv.Itoa(len(perms))+" perms...\n", use_color)
	}

	if (domain != "") || (buckets_file != "") {
		// Create n concurrent workers
		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				// load config for anonymous access
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("us-east-1"), config.WithCredentialsProvider(aws.AnonymousCredentials{}))
				if err != nil {
					log.Fatal(err)
				}

				cfg.Credentials = nil
				client := s3.NewFromConfig(cfg, func(o *s3.Options) {
					o.UsePathStyle = false
				})

				for bucket := range buckets_c { // receive bucket from buckets channel
					format_str := bucket

					region, err := manager.GetBucketRegion(context.TODO(), client, bucket) // try to get bucket region
					if err != nil {                                                        // handle error
						var bnf manager.BucketNotFound

						if errors.As(err, &bnf) { // check if err means that bucket wasn't found
							if use_color {
								fmt.Println(bucket, "| "+cyan("STATUS")+": "+red("Not Found"))
							} else {
								fmt.Println(bucket, "| STATUS: Not Found")
							}
							continue // continue to next bucket since it doesn't exists
						}

						fmt.Println("error:", err)
						continue
					}

					if output != "" {
						_, err = out_f.WriteString(string(bucket))
						if err != nil {
							log.Fatal(err)
						}
					}

					found_buckets = append(found_buckets, bucket)

					if use_color {
						format_str = format_str + " | " + cyan("STATUS") + ": " + green("Found") + " | " + cyan("REGION") + ": " + green(region)
					} else {
						format_str = format_str + " | STATUS: Found" + " | REGION: " + region
					}

					GetBucketACLInput := &s3.GetBucketAclInput{
						Bucket: aws.String(bucket),
					}

					// try to retrieve ACLs
					GetBucketACLOutput, err := client.GetBucketAcl(context.TODO(), GetBucketACLInput)
					if err != nil { // handle error
						if use_color {
							format_str = format_str + " | " + cyan("GET ACL") + ": " + red("Failed")
						} else {
							format_str = format_str + " | GET ACL: Failed"
						}
					} else {
						groups := map[string]string{
							"http://acs.amazonaws.com/groups/global/AllUsers":           "Everyone",
							"http://acs.amazonaws.com/groups/global/AuthenticatedUsers": "Authenticated AWS users",
						}

						permissions := map[string][]string{}

						// iterate over ACLs
						for _, grant := range GetBucketACLOutput.Grants {
							if grant.Grantee.Type == "Group" {
								for group := range groups {
									if *grant.Grantee.URI == group {
										permissions[groups[group]] = append(permissions[groups[group]], string(grant.Permission))
									}
								}
							}
						}

						ACL := []string{}

						for permission := range permissions {
							ACL = append(ACL, fmt.Sprintf("%s: %s", permission, strings.Join(permissions[permission], ", ")))
						}

						if use_color {
							format_str = fmt.Sprintf(format_str+" | "+cyan("GET ACL")+": "+green("%s"), ACL)
						} else {
							format_str = fmt.Sprintf(format_str+" | GET ACL: %s", ACL)
						}
					}

					PutObjectInput := &s3.PutObjectInput{
						Bucket: aws.String(bucket),
						Key:    aws.String("d3ext.txt"),
					}

					// try to upload a file
					_, err = client.PutObject(context.TODO(), PutObjectInput)
					if err != nil { // handle error
						if use_color {
							format_str = format_str + " | " + cyan("PUT OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | PUT OBJECTS: Failed"
						}
					} else {
						if use_color {
							format_str = format_str + " | " + cyan("PUT OBJECTS") + ": " + green("Success")
						} else {
							format_str = format_str + " | PUT OBJECTS: Success"
						}
					}

					ListObjectsV2Input := &s3.ListObjectsV2Input{
						Bucket: aws.String(bucket),
					}

					// try to list files and folders
					_, err = client.ListObjectsV2(context.TODO(), ListObjectsV2Input)
					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | LIST OBJECTS: Failed"
						}
					} else {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + green("Success")
						} else {
							format_str = format_str + " | LIST OBJECTS: Success"
						}
					}

					// finally print results
					fmt.Println(format_str)
				}

				wg.Done()
			}()
		}

		if domain != "" {
			if (!strings.HasPrefix(domain, "http://")) || (!strings.HasPrefix(domain, "https://")) {
				domain = "https://" + domain
			}

			parse, err := tld.Parse(domain)
			if err != nil {
				log.Fatal(err)
			}

			buckets_c <- parse.Domain + "." + parse.TLD // https://example.com.s3.amazonaws.com
			buckets_c <- parse.Domain                   // https://example.s3.amazonaws.com
			buckets_c <- parse.Domain + "-" + parse.TLD // https://example-com.s3.amazonaws.com
			for _, p := range perms {
				buckets_c <- p + "-" + parse.Domain + "." + parse.TLD // https://<perm>-example.com.s3.amazonaws.com
				buckets_c <- p + "-" + parse.Domain                   // https://<perm>-example.s3.amazonaws.com
				buckets_c <- parse.Domain + "-" + p + "." + parse.TLD // https://example-<perm>.com.s3.amazonaws.com
				buckets_c <- parse.Domain + "-" + p                   // https://example-<perm>.s3.amazonaws.com
			}

		} else if buckets_file != "" {
			f, err := os.Open(buckets_file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				bucket := scanner.Text()
				buckets_c <- bucket
			}
		}

		close(buckets_c)
		wg.Wait()

	} else if (list != "") || (stdin) {

		var f *os.File
		var err error

		if list != "" {
			f, err = os.Open(list)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

		} else if stdin {
			f = os.Stdin
		}

		// create n concurrent workers
		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("us-east-1"), config.WithCredentialsProvider(aws.AnonymousCredentials{}))
				if err != nil {
					log.Fatal(err)
				}

				cfg.Credentials = nil
				client := s3.NewFromConfig(cfg, func(o *s3.Options) {
					o.UsePathStyle = false
				})

				for bucket := range buckets_c { // receive bucket from buckets channel
					format_str := bucket

					region, err := manager.GetBucketRegion(context.TODO(), client, bucket)
					if err != nil {
						var bnf manager.BucketNotFound

						if errors.As(err, &bnf) {
							if use_color {
								fmt.Println(bucket, "| "+cyan("STATUS")+": "+red("Not Found"))
							} else {
								fmt.Println(bucket, "| STATUS: Not Found")
							}
							continue
						}

						fmt.Println("error:", err)
						continue
					}

					if output != "" {
						_, err = out_f.WriteString(string(bucket))
						if err != nil {
							log.Fatal(err)
						}
					}

					found_buckets = append(found_buckets, bucket)

					if use_color {
						format_str = format_str + " | " + cyan("STATUS") + ": " + green("Found") + " | " + cyan("REGION") + ": " + green(region)
					} else {
						format_str = format_str + " | STATUS: Found" + " | REGION: " + region
					}

					GetBucketACLInput := &s3.GetBucketAclInput{
						Bucket: aws.String(bucket),
					}

					GetBucketACLOutput, err := client.GetBucketAcl(context.TODO(), GetBucketACLInput)
					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("GET ACL") + ": " + red("Failed")
						} else {
							format_str = format_str + " | GET ACL: Failed"
						}
					} else {
						groups := map[string]string{
							"http://acs.amazonaws.com/groups/global/AllUsers":           "Everyone",
							"http://acs.amazonaws.com/groups/global/AuthenticatedUsers": "Authenticated AWS users",
						}

						permissions := map[string][]string{}

						for _, grant := range GetBucketACLOutput.Grants {
							if grant.Grantee.Type == "Group" {
								for group := range groups {
									if *grant.Grantee.URI == group {
										permissions[groups[group]] = append(permissions[groups[group]], string(grant.Permission))
									}
								}
							}
						}

						ACL := []string{}

						for permission := range permissions {
							ACL = append(ACL, fmt.Sprintf("%s: %s", permission, strings.Join(permissions[permission], ", ")))
						}

						if use_color {
							format_str = fmt.Sprintf(format_str+" | "+cyan("GET ACL")+": "+green("%s"), ACL)
						} else {
							format_str = fmt.Sprintf(format_str+" | GET ACL: %s", ACL)
						}
					}

					PutObjectInput := &s3.PutObjectInput{
						Bucket: aws.String(bucket),
						Key:    aws.String("d3ext.txt"),
					}

					_, err = client.PutObject(context.TODO(), PutObjectInput)
					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("PUT OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | PUT OBJECTS: Failed"
						}
					} else {
						if use_color {
							format_str = format_str + " | " + cyan("PUT OBJECTS") + ": " + green("Success")
						} else {
							format_str = format_str + " | PUT OBJECTS: Success"
						}
					}

					ListObjectsV2Input := &s3.ListObjectsV2Input{
						Bucket: aws.String(bucket),
					}

					_, err = client.ListObjectsV2(context.TODO(), ListObjectsV2Input)
					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | LIST OBJECTS: Failed"
						}
					} else {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + green("Success")
						} else {
							format_str = format_str + " | LIST OBJECTS: Success"
						}
					}

					fmt.Println(format_str)
				}

				wg.Done()
			}()
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			domain := scanner.Text()

			if (!strings.HasPrefix(domain, "http://")) || (!strings.HasPrefix(domain, "https://")) {
				domain = "https://" + domain
			}

			parse, err := tld.Parse(domain)
			if err != nil {
				log.Fatal(err)
			}

			buckets_c <- parse.Domain + "." + parse.TLD // https://example.com.s3.amazonaws.com
			buckets_c <- parse.Domain                   // https://example.s3.amazonaws.com
			buckets_c <- parse.Domain + "-" + parse.TLD // https://example-com.s3.amazonaws.com
			for _, p := range perms {
				buckets_c <- p + "-" + parse.Domain + "." + parse.TLD // https://<perm>-example.com.s3.amazonaws.com
				buckets_c <- p + "-" + parse.Domain                   // https://<perm>-example.s3.amazonaws.com
				buckets_c <- parse.Domain + "-" + p + "." + parse.TLD // https://example-<perm>.com.s3.amazonaws.com
				buckets_c <- parse.Domain + "-" + p                   // https://example-<perm>.s3.amazonaws.com
			}
		}
	}

	if json_output != "" {
		json_buckets := BucketsInfo{
			Buckets: found_buckets,
			Length:  len(found_buckets),
			Time:    core.TimerDiff(t1).String(),
		}

		json_body, err := json.Marshal(json_buckets)
		if err != nil {
			log.Fatal(err)
		}

		_, err = out_f.WriteString(string(json_body))
		if err != nil {
			log.Fatal(err)
		}
	}

	if !quiet {
		if counter >= 1 {
			if use_color {
				fmt.Println("\n["+green("+")+"]", counter, "buckets found!")
			} else {
				fmt.Println("\n[+]", counter, "buckets found!")
			}
		}

		if output != "" {
			if counter >= 1 {
				if use_color {
					fmt.Println("["+green("+")+"] Buckets written to", output)
				} else {
					fmt.Println("[+] Buckets written to", output)
				}
			}
		}

		if use_color {
			if output != "" || counter >= 1 {
				fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			} else {
				fmt.Println("\n["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
			}
		} else {
			if output != "" || counter >= 1 {
				fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
			} else {
				fmt.Println("\n[+] Elapsed time:", core.TimerDiff(t1))
			}
		}
	}
}
