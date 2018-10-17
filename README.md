## What is Picolo?

Picolo is a database network built for web 3.0. It gives developers a new way of storing data on a decentralized network
of nodes clost to their users.
- [Status](#status)
- [Quickstart](#quickstart)
- [Run a node](#run-a-node)
- [Need Help?](#need-help)
- [Roadmap](#roadmap)

## Status

Picolo is in prototype stage. See our [Roadmap](#roadmap) for a list of features planned or in development.

## Quickstart

- Install dependencies

```
npm install pg request async
```
- Create an app
```javascript
const pg = require('pg')
const request = require('request')

request.post('https://picolo.app/create', { json: { name: 'testApp' } },
    (err, res, body) => {
        console.log(err, body)
    }
)
```
- Connect to an app
```javascript
const pg = require('pg')
const request = require('request')

let pool = new pg.Pool()
request.get('https://picolo.app/testApp',
    (err, res, body) => {
        console.log(err, body)
        if (!err && res.statusCode == 200) {
            pool = new pg.Pool({connectionString: body,})
        }
    }
)
```
- Run queries
```javascript
const pg = require('pg')
const request = require('request')

// Create a table
pool.query('CREATE TABLE IF NOT EXISTS t3 (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name STRING)', (err, res) => {
    console.log(err, res)
})

// Insert some data
pool.query("INSERT INTO t3 (name) VALUES ('picolo')", (err, res) => {
    console.log(err, res)
})

// Query it
pool.query('SELECT name FROM t3', (err, res) => {
    console.log(err, res)
    res.rows.forEach(function(row) {
        console.log(row)
    })
})

pool.end()
```
Full example is [here](https://github.com/picolonet/picolo-examples/blob/master/nodejs/index.js)

## Run a node

Picolo is a decentralized network with an open participation model. Anyone can run a node and
contribute to the capacity of the network.

- Install on linux
```
curl -L https://github.com/picolonet/cockroach/releases/download/1.0.0/picolo-linux-amd64.tar.gz | tar xvz
sudo cp -i picolo /usr/local/bin
```
- Install on mac
```
curl -L https://github.com/picolonet/cockroach/releases/download/1.0.0/picolo-darwin-amd64.tar.gz | tar xvz
sudo cp -i picolo /usr/local/bin
```
- Run
```
picolo
```

## Need Help?

- [Join us on Gitter](https://gitter.im/picolonet/general) - This is the most immediate
  way to connect with Picolo engineers.

## Roadmap

- Replace RAFT consensus with a version that is tolerant to Byzantine faults
- Replace network layer with an optimized DHT based layer
- Replace discovery service with the network layer
