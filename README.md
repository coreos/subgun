# subgun

Subscribe to a mailgun backed mailing list via a web interface.

## Build and Configuration

Build subgun using the build script:

```
./build
```

Configure subgun by setting a series of environment variables:

```
export SUBGUN_LISTEN=127.0.0.1:8080
export SUBGUN_LISTS=<mailgun list 1>,<mailgun list 2>,...
export SUBGUN_API_KEY=<mailgun api key>
```

Once configured, you can simply run the binary:

```
./subgun
```

## Subscription Workflow

POST to the subscribe URL. See `subscribe.html` or do it with curl:

```
curl http://localhost:8080/subscribe/etcd-dev@lists.coreos.com -d "email=brandon@example.com"
```

This will email `brandon@example.com` with a confirmation email with a URL that
contains a secret token. Clicking on the link will change the subscription
status to "Subscribed".

## Unsubscribe Workflow

For now use the native workflow in Mailgun.

## What about archives?

http://gmane.org/subscribe.php

## TODO

- Already subscribed user error message instead of user error
- Cleanup HTTP handling
