# cf willitconnect

[![wercker status](https://app.wercker.com/status/6947d629f21a39d5700a5c50b705fe86/s "wercker status")](https://app.wercker.com/project/bykey/6947d629f21a39d5700a5c50b705fe86)

Cloud Foundry CLI plugin to validate connectivity between a CF instance and a thing

This plugin makes it even easier to use [Willitconnect](https://github.com/krujos/willitconnect) to validate if you can connect from CF to a thing.

This simplifies troubleshooting, but was also partly inspired by [lmgtfy.com](http://lmgtfy.com/)

##Usage

Default functionality assumes willitconnect is the route used for the willitconnect application running on your CF instance, and creates a socket connection to the specified port and host.  If desired, you can specify an alternate
route for willitconnect and/or specify a proxy for willitconnect to use.    In addition, if the host you pass is a url,
willitconnect will attempt an http connection

```
$ cf willitconnect -host=<host> -port=<port>
$ cf willitconnect <url>
$ cf willitconnect -host=<host> -port=<port> -proxyHost=<proxyHost> -proxyPort=<proxyPort>
```

##install

```
$ cf install-plugin cf-willitconnect -r CF-Community
```

##Todo

* Add detailed http status info
