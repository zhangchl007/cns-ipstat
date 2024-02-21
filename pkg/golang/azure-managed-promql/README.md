## demo for azure managed prometheus promql by restapi

```bash
export AUTH_ENDPOINT=https://login.microsoftonline.com/{Tenenet ID}/oauth2/token
export CLIENT_ID=xxxxx
export CLIENT_SECRET=xxxxx

go run main.go
go build -o promqlquery
./promqlquery
```

