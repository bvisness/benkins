Design:

- Server/runner architecture like Gitlab CI
- Runners can either poll for new work or maintain a socket connection (only polling for MVP)
- No running on the server program directly, but of course the server computer could also be a runner
- Server can poll repo for new jobs, or listen to webhooks (only polling for MVP)
- Server has webserver for viewing jobs and results

All communication happens over sockets because we need real-time streaming data stuff. We need some kind of format
to use for this communication, so let's invent something:

```
Route=routeName
AuthKey=AXBYCZ
Timeout=3000
::::
This is the body. The body will go on forever and ever and ever. That is, until you hit the final delimiter.
```

STTP is designed to deal with the "one-way" nature of the server/runner architecture. Runners are not publicly
available on the internet, so the runner instead establishes a connection to the server and holds it open.

STTP uses one socket connection per runner, called the initiator, to handle requests either from runner to server or
server to runner. The initiator sends only header information. Regardless of  
header information. The runner receives this header information and establishes a new socket connection to handle
that request in isolation.

 

```
Route=routeName
AuthKey=AXBYCZ
Timeout=3000
::::
This is the body. The body will go on forever and ever and ever. That is, until you hit the final delimiter.
```