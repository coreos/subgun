# subscribegun

Subscribe to a mailing list on mailgun via a web interface

## Subscription Workflow

curl -s --user 'api:key-3ax6xnjp29jd6fds4gc373sgvjxteol0' \
    https://api.mailgun.net/v2/lists/dev@samples.mailgun.org/members \
    -F subscribed=False \
    -F address='bar@example.com' \
    -F name='Bob Bar' \
    -F vars='{"confirmationCode": "aSecret"}'

{
  "member": {
      "vars": {
          "confirmationCode": "aSecret"
      },
      "name": "Bob Bar",
      "subscribed": true,
      "address": "bar@example.com"
  },
  "message": "Mailing list member has been created"
}

## What about archives?

http://gmane.org/subscribe.php
