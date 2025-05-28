# API Tester CLI

## Features

- ✅ Send GET/POST/PUT/DELETE requests
- 💾 Save and reuse API test cases
- 🎨 JSON pretty-printing for responses
- 📦 Single binary - no dependencies
- 🔐 Config stored in `~/.apitester/`
- 🚀 Fast and scriptable


# Basic Commands
## Send a GET request
apitester send -u https://api.example.com/users
## Send a POST request with JSON
apitester send -X POST -u https://api.example.com/users \
  -d '{"name": "John"}' \
  -H "Content-Type=application/json"
## Save a request
apitester save get-users -X GET -u https://api.example.com/users
## Run a saved request
apitester run get-users
## List saved requests
apitester list
