# token_notifier
Notify me via Discord webhook when WoW token value exceeds a predefined threshold.

## Use
You'll need to set some enviroment variables for this to run:
```
TKNT_BLIZZARD_CLIENTID=YOUR_BLIZZARD_CLIENT_ID
TKNT_BLIZZARD_CLIENTSECRET=YOUR_BLIZZARD_CLIENT_SECRET
TKNT_DISCORD_WEBHOOK=YOUR_DISCORD_WEBHOOK_URL
TKNT_NOTIFICATION_THRESHOLD=3000000000
```

Blizzard returns the token price in copper, so this number is 10,000 times larger than you might expect. Tack an extra 4 0s on to the end of your target number and you'll have the right threshold. The example above equates to 300k.

You can set these in whatever you're using to schedule tasks, or you can set them in a .env file.
