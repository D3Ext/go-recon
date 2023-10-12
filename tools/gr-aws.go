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
  INPUT:
    -d, -domain string          target domain to find its S3 AWS buckets and their info (i.e. example.com)
    -l, -list string            file containing a list of domains to find their S3 AWS buckets and their info (one domain per line)
    -bl, -bucket-list string    file containing a list of bucket names to check and list info (one name per line)

  PERMUTATIONS:
    -pl, -perms-list string       file containing a list of permutations, if not especified, default perms are used (doesn't apply with -b)
    -level int                    level of permutations to generate with (-d) and (-l) params (1-5, default=3)
    -f, -format string            pattern to generate bucket names based on permutations (default: <perm>.<domain>.<tld>)
    -all-formats                  use 4 most common formats to generate bucket names (slow)
    -lf, -list-formats            show allowed variables that will be replaced by corresponding values

  OUTPUT:
    -o, -output string          file to write buckets urls into
    -oj, -output-json string    file to write buckets urls into (JSON format)

  CONFIG:
    -p, -proxy string     proxy to send requests through (i.e. http://127.0.0.1:8080)
    -t, -timeout int      milliseconds to wait before each request timeout (default=5000)
    -w, -workers int      number of concurrent workers (default=10)
    -c, -color            print colors on output (recommended)
    -q, -quiet            don't print banner, only output

  DEBUG:
    -version        show go-recon version
    -h, -help       print help panel
  
Examples:
    gr-aws -d example.com -o buckets.txt -c
    gr-aws -l subdomains.txt -pl perms.txt
    gr-aws -d example.com -level 4 -w 5
    gr-aws -bl buckets_to_check.txt
    cat domains.txt | gr-aws -c
    `)
}

// nolint: gocyclo
func main() {
	var domain string
	var list string
	var buckets_list string
	var perms_list string
	var level int
	var format string
	var allformats bool
	var list_formats bool
	var proxy string
	var timeout int
	var workers int
	var output string
	var json_output string
	var quiet bool
	var use_color bool
	var version bool
	var help bool
	var stdin bool

	flag.StringVar(&domain, "d", "", "")
	flag.StringVar(&domain, "domain", "", "")
	flag.StringVar(&list, "l", "", "")
	flag.StringVar(&list, "list", "", "")
	flag.StringVar(&buckets_list, "bl", "", "")
	flag.StringVar(&buckets_list, "buckets-list", "", "")
	flag.StringVar(&perms_list, "pl", "", "")
	flag.StringVar(&perms_list, "perms-list", "", "")
	flag.IntVar(&level, "level", 3, "")
	flag.StringVar(&format, "f", "", "")
	flag.StringVar(&format, "format", "", "")
	flag.BoolVar(&allformats, "all-formats", false, "")
	flag.BoolVar(&list_formats, "lf", false, "")
	flag.BoolVar(&list_formats, "list-formats", false, "")
	flag.IntVar(&workers, "w", 10, "")
	flag.IntVar(&workers, "workers", 10, "")
	flag.StringVar(&proxy, "p", "", "")
	flag.StringVar(&proxy, "proxy", "", "")
	flag.IntVar(&timeout, "t", 5000, "")
	flag.IntVar(&timeout, "timeout", 5000, "")
	flag.StringVar(&output, "o", "", "")
	flag.StringVar(&output, "output", "", "")
	flag.StringVar(&json_output, "oj", "", "")
	flag.StringVar(&json_output, "output-json", "", "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&use_color, "c", false, "")
	flag.BoolVar(&use_color, "color", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()

	t1 := core.StartTimer()

	if version {
		fmt.Println("go-recon version:", core.Version())
		os.Exit(0)
	}

	if !quiet {
		fmt.Println(core.Banner())
	}

	if help {
		helpPanel()
		os.Exit(0)
	}

	if list_formats {
		listFormats(quiet)
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
	if (domain == "") && (list == "") && (buckets_list == "") && (!stdin) {
		helpPanel()
		os.Exit(0)
	}

	if ((domain != "") && (list != "")) || ((domain != "") && (buckets_list != "")) || ((list != "") && (buckets_list != "")) {
		helpPanel()
		core.Red("You can't use (-d), (-l) or (-b) at same time", use_color)
		os.Exit(0)
	}

	if (format != "") && (allformats) {
		helpPanel()
		core.Red("You can't use (-f) and (-all-formats) at same time", use_color)
		os.Exit(0)
	}

	if level != 1 && level != 2 && level != 3 && level != 4 && level != 5 {
		helpPanel()
		core.Red("Invalid level! Allowed values: 1-5", use_color)
		os.Exit(0)
	}

	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}

	if format == "" {
		format = "<perm>.<domain>.<tld>"
	}

	// define variables which will be used to write buckets to output files
	var txt_out *os.File
	if output != "" {
		txt_out, err = os.Create(output)
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

			json_out, err := os.Create(json_output)
			if err != nil {
				log.Fatal(err)
			}

			_, err = json_out.WriteString(string(json_body))
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
		perms = core.GetPerms(level) // Default permutations list (285 words)
	}

	client := core.DefaultHttpClient()

	var counter int
	var wg sync.WaitGroup
	buckets_c := make(chan string) // Create channel and wait group for concurrency

	if !quiet {
		core.Warning("Use with caution.", use_color)
		core.Magenta("Concurrent workers: "+strconv.Itoa(workers), use_color)
		if proxy != "" {
			core.Magenta("Proxy: "+proxy, use_color)
		}
		core.Magenta("Looking for S3 buckets with "+strconv.Itoa(len(perms))+" perms...\n", use_color)
	}

	if (domain != "") || (buckets_list != "") {
		// Create n concurrent workers
		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func() {
				// load config for anonymous access
				cfg, err := config.LoadDefaultConfig(
					context.TODO(),
					config.WithDefaultRegion("us-east-1"),
					config.WithCredentialsProvider(aws.AnonymousCredentials{}),
					config.WithHTTPClient(client),
				)
				if err != nil {
					log.Fatal(err)
				}

				cfg.Credentials = nil

				for bucket := range buckets_c { // receive bucket from buckets channel
					format_str := bucket

					region, err := manager.GetBucketRegion(context.TODO(), s3.NewFromConfig(cfg), bucket) // try to get bucket region
					if err != nil {                                                                       // handle error
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
						_, err = txt_out.WriteString(string(bucket))
						if err != nil {
							log.Fatal(err)
						}
					}

					found_buckets = append(found_buckets, bucket)
					counter += 1

					if use_color {
						format_str = format_str + " | " + cyan("STATUS") + ": " + green("Found") + " | " + cyan("REGION") + ": " + green(region)
					} else {
						format_str = format_str + " | STATUS: Found" + " | REGION: " + region
					}

					client := s3.NewFromConfig(cfg, func(o *s3.Options) { // create client for each iteration so region is correct
						o.Region = region
						o.UsePathStyle = false
					})

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
					output, err := client.ListObjectsV2(context.TODO(), ListObjectsV2Input)
					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | LIST OBJECTS: Failed"
						}
					} else {

						var obj_counter int = 0
						for i := 0; i < len(output.Contents); i++ {
							obj_counter += 1
						}

						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + green("Success") + " (" + cyan(strconv.Itoa(obj_counter)) + " objects)"
						} else {
							format_str = format_str + " | LIST OBJECTS: Success (" + strconv.Itoa(obj_counter) + " objects)"
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

			if allformats {
				for _, p := range perms {
					buckets_c <- p + "-" + parse.Domain + "." + parse.TLD // https://<perm>-example.com.s3.amazonaws.com
					buckets_c <- p + "-" + parse.Domain                   // https://<perm>-example.s3.amazonaws.com
					buckets_c <- parse.Domain + "-" + p + "." + parse.TLD // https://example-<perm>.com.s3.amazonaws.com
					buckets_c <- parse.Domain + "-" + p                   // https://example-<perm>.s3.amazonaws.com
				}
			} else {
				for _, p := range perms {
					n := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(format, "<tld>", parse.TLD), "<domain>", parse.Domain), "<perm>", p)
					buckets_c <- n
				}
			}

		} else if buckets_list != "" {
			f, err := os.Open(buckets_list)
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
				cfg, err := config.LoadDefaultConfig(
					context.TODO(),
					config.WithDefaultRegion("us-east-1"),
					config.WithCredentialsProvider(aws.AnonymousCredentials{}),
					config.WithHTTPClient(client),
				)
				if err != nil {
					log.Fatal(err)
				}

				cfg.Credentials = nil

				for bucket := range buckets_c { // receive bucket from buckets channel
					format_str := bucket

					region, err := manager.GetBucketRegion(context.TODO(), s3.NewFromConfig(cfg), bucket)
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
						_, err = txt_out.WriteString(string(bucket))
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

					client := s3.NewFromConfig(cfg, func(o *s3.Options) { // create client for each iteration so region is correct
						o.Region = region
						o.UsePathStyle = false
					})

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
							format_str = format_str + " | PUT OBJECTS: Success "
						}
					}

					ListObjectsV2Input := &s3.ListObjectsV2Input{
						Bucket: aws.String(bucket),
					}

					output, err := client.ListObjectsV2(context.TODO(), ListObjectsV2Input)

					if err != nil {
						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + red("Failed")
						} else {
							format_str = format_str + " | LIST OBJECTS: Failed"
						}
					} else {

						var obj_counter int = 0
						for i := 0; i < len(output.Contents); i++ {
							obj_counter += 1
						}

						if use_color {
							format_str = format_str + " | " + cyan("LIST OBJECTS") + ": " + green("Success") + " (" + cyan(strconv.Itoa(obj_counter)) + " objects)"
						} else {
							format_str = format_str + " | LIST OBJECTS: Success (" + strconv.Itoa(obj_counter) + " objects)"
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

		json_out, err := os.Create(json_output)
		if err != nil {
			log.Fatal(err)
		}

		_, err = json_out.WriteString(string(json_body))
		if err != nil {
			log.Fatal(err)
		}
	}

	if !quiet {
		fmt.Println()
		if counter >= 1 {
			if output != "" {
				core.Green("Buckets written to "+output+" (TXT)", use_color)
			}

			if json_output != "" {
				core.Green("Buckets written to "+json_output+" (JSON)", use_color)
			}

			core.Green(strconv.Itoa(counter)+" buckets found", use_color)

		} else {
			core.Red("No buckets found", use_color)
		}

		if use_color {
			fmt.Println("["+green("+")+"] Elapsed time:", green(core.TimerDiff(t1)))
		} else {
			fmt.Println("[+] Elapsed time:", core.TimerDiff(t1))
		}
	}
}

func listFormats(quiet bool) {
	if !quiet {
		fmt.Println("[*] Available variables:")
		fmt.Println("\t<perm>      placeholder to replace with all permutations")
		fmt.Println("\t<domain>    placeholder to replace with domain")
		fmt.Println("\t<tld>       placeholder to replace with domain TLD")
		fmt.Println("\nExample: <perm>-<domain>.<tld>")
	} else {
		fmt.Println("<perm>\n<domain>\n<tld>")
	}
}
