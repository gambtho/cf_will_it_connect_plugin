# cf willitconnect

[![wercker status](https://app.wercker.com/status/e1f851f98e9028a7ccd9a8a38e84eca0/m "wercker status")](https://app.wercker.com/project/bykey/e1f851f98e9028a7ccd9a8a38e84eca0)

Cloud Foundry CLI plugin to validate connectivity between a CF instance and a thing

This plugin makes it even easier to use [Willitconnect](https://github.com/krujos/willitconnect) to validate if you can connect from CF to a thing.

This simplifies troubleshooting, but was also partly inspired by [lmgtfy.com](http://lmgtfy.com/)

##Usage

Initial release provides basic socket connectivity check, and assumes willitconnect is the route used for the willitconnect application running on your CF instance

```
$ cf willitconnect <host> <port>
```

##Todo

* Add http status info
* Allow proxy usage
* Adjust willitconnect route
