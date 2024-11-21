from flask import Flask, request

app = Flask(__name__)

@app.route('/')
def home():
    name = request.args.get('name', 'anonymous')
    # Directly include the user input in the response
    return f"<p>Hello {name}</p>"

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)

