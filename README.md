## DNSBrute

#### Feature

- query over api
    - http://www.hackertarget.com/
    - ...
- dict based

#### Advantage

- Fast: 5000~10000+ domains /sec, depending on the network
- Pan-DNS identification

#### Usage

```
➜  dnsbrute ./dnsbrute
Usage:
  ./dnsbrute [Options]

Options
  -debug
    	Show debug information
  -dict string
    	Dict file (default "dict/53683.txt")
  -domain string
    	Domain to brute
  -rate int
    	Transmit rate of packets (default 10000)
  -retry int
    	Limit for retry (default 3)
  -server string
    	Address of DNS server (default "8.8.8.8:53")
  -version
    	Show program's version number and exit
```

#### Screenshot

![Screenshot](screenshot.png)

#### Depends

- go: 1.10
- packages: install with `glide`
