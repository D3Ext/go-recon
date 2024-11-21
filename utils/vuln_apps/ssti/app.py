from flask import Flask, request, render_template_string
import os

app = Flask(__name__)

@app.route("/")
def index():
    name = request.args.get('name', 'anonymous')
    template = f"Hello {name}\n"
    return render_template_string(template, name=name)

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=80)
