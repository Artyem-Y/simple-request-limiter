# Simple Request Limiter

## Installation

1. Clone this repository

2. Run `go mod download`

## Example

### Run example

Try to hit this endpoint: `POST` http://localhost:8080/urls

with request

```json
{
    "urls": "https://google.com\r\nhttps://google.com\r\nhttps://google.com\r\nhttps://google.com\r\nhttps://google.com"
}
```

and you will get this response

```json
{
    "data": "17cdc549e3a685778c55026cd29a9e4a13d4aaa1\r\n453df402bcfb8c64704dd1eb5cc62e3288d9d751\r\nd9d5871b2baec3da9f7e62b1eb68e1b5ac27ba33\r\nba666440b85ecdbedf1abee912d7c07f79d0713b\r\n4ada1814d7e627e60665011f361a5b7acae82abd",
    "message": "Status OK"
}
```

If your request got a limitation, you will get the "too many request" response by default.

### Custom Request Limiter

Limit all incoming request

```go
reqIPLimiter := limiter.NewRequestLimitService(10*time.Second, 100, nil)

http.HandleFunc("/resources", reqIPLimiter.Limit(resourcesHandler))
```

You can write your own limiter function. Here is example to limit request per IP

```go
reqIPLimiter := limiter.NewRequestLimitService(10*time.Second, 100, func(r *http.Request) string {
    return r.RemoteAddr
})

http.HandleFunc("/resources", reqIPLimiter.Limit(resourcesHandler))
```

Another example is limit request by session id that stored in cookie:

```go
reqSessionLimiter := limiter.NewRequestLimitService(10*time.Second, 100, func(r *http.Request) string {
    return r.Cookie("Session-ID")
})

http.HandleFunc("/resources", reqIPLimiter.Limit(resourcesHandler))
```

### Custom Blocked Request Handler

You can use your own response handler when current request is blocked

```go
limiter.OnTooManyRequest(func(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "please try again later", http.StatusTooManyRequest)
})
```
