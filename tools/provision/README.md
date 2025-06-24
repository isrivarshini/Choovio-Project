# Mainflux Things and Channels Provisioning Tool

A simple utility to create a list of channels and things connected to these channels with possibility to create certificates for mTLS use case.

This tool is usefule for testing, and it creates a TOML format output (on stdout, can be redirected into the file of needed)
that can be used by Mainflux MQTT benchmarking tool (`mqtt-bench`).

### Usage
```
./provision --help
Tool for provisioning series of Mainflux channels and things and connecting them together.
Complete documentation is available at https://mainflux.readthedocs.io

Usage:
  provision [flags]

Flags:
      --ca string         CA for creating and signing things certificate (default "ca.crt")
      --cakey string      ca.key for creating and signing things certificate (default "ca.key")
  -h, --help              help for provision
      --host string       address for mainflux instance (default "https://localhost")
      --num int           number of channels and things to create and connect (default 10)
  -p, --password string   mainflux users password
      --ssl               create certificates for mTLS access
  -u, --username string   mainflux user
```

Example:
``` 
./provision -u mirkot@mainflux.com -p test1234 --host https://142.93.118.47
```

If you want to create a list of channels with certificates:

```
./provision --host http://localhost --num 10 -u test@mainflux.com -p test1234 --ssl true --ca ../../docker/ssl/certs/ca.crt --cakey ../../docker/ssl/certs/ca.key

```

>`ca.crt` and `ca.key` are used for creating things certificate and for HTTPS,
> if you are provisioning on remote server you will have to get these files to your local 


Example of output:

```
[[mainflux]]
  ChannelID = "42053920-439a-461f-bd81-dece1ef06089"
  ThingID = "49f3c38c-400e-4171-a954-24850150124b"
  ThingKey = "835d74ab-7232-49b7-bc6d-a1ca97c6eecc"
  MTLSCert = "-----BEGIN CERTIFICATE-----\nMIIFmDCCA4CgAwIBAgIQVvvHUH9XPYMGF2F/YzfxFzANBgkqhkiG9w0BAQsFADBX\nMRIwEAYDVQQDDAlsb2NhbGhvc3QxETAPBgNVBAoMCE1haW5mbHV4MQwwCgYDVQQL\nDANJb1QxIDAeBgkqhkiG9w0BCQEWEWluZm9AbWFpbmZsdXguY29tMB4XDTE5MDgw\nNzIwMzU0NVoXDTE5MDgwNzIwMzU0NVowVTERMA8GA1UEChMITWFpbmZsdXgxETAP\nBgNVBAsTCG1haW5mbHV4MS0wKwYDVQQDEyQ4MzVkNzRhYi03MjMyLTQ5YjctYmM2\nZC1hMWNhOTdjNmVlY2MwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDC\nhC/TAiiE2OKug73B5lWxUfzpUfukIIl+Ge8Xz6CIqzGlvPoggpCm5igZUC5YTpg4\neIC550kyq8pLDZzck/f7Qq0aJq3wBYLtA64XzfTrG84MI+yt3mniKYmZQInW1GXC\nStS+jzAnhTPr5urnMaPLm5g9UTlvMN6TwRMlf4kSlSCR0KN0kKH6yRU9MEEwvt+2\nTu5RWLN1LwZzG5ud0AAjT5CL08akloUpSEt1KLY2c7uf9l/1upkxBmhKoSs0sRW0\n25aX6c0qLOnKDRrsUKI8YbYkKAKsRtGRPl9q1j/Im4mE/9EwGgibdwVItz2Pmq/k\nOQjZ+f3tEKXwRD3yK9Aas1u/K5tf35aK5Eznr1f2FyM8uQ2SyEcXVikYA32hoI6W\nQRkTTlueG4K2H/KN2KL5MC+N7h17Bw5ZCS0201xnhDLjTaeGFr0SibFHJA8IVOeL\npTKohB4r1gAOKqzIQ3IYdJ/3jk31SxGw7Et3QBs+wxyUHLCalBFIk4R+GC276Buk\nFR3fFOIsm5+f5udZTYrikW0dWXHxhhSOFlTurQP5aS0K5HpdwnOkMWTKYMwmwHut\nxF9ycPc5gCjBC8wPGbrCkA9XJdKw+VTySg6tTgEnyLvkccdi07Fs13qQcT1MK2sv\nf9qEVdVBRnsB1i2vffHRF3j5RGK1hbKO1AX4fICJMQIDAQABo2IwYDAOBgNVHQ8B\nAf8EBAMCB4AwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMA4GA1UdDgQH\nBAUBAgMEBjAfBgNVHSMEGDAWgBRs4xR91qEjNRGmw391xS7x6Tc+8jANBgkqhkiG\n9w0BAQsFAAOCAgEAcZjTZLAPOr9Xm7ASxieAj5pdoRKz5Q3y4kZ2oCJ/BtIfN25C\nTEuRvNLkLZHZo/ZeOU0XyNfpHiOqc0J4R2zgrpdncH0+viC22bhcec95+fNrES+G\nGpbw0BufFPcn0cDeJQbtg//4eeCiXaNZgui8Og9t2D4TgYPwPmi4l6tYXDpMUaT0\n5Vrr6VXezbHUpDtSaG2aoo4Z7nrEhtXKnYlpo8htEbCH8S3zxbmWx3bn8gS8Dsr9\nSyqSxiAMYjIBxwCmgTi6iaINmSHxtMhqU5mmPVBa1vZBiNBQlvW2Yslvl0aMdgVD\nLzz8JULXBuKwIdHAbmrDgXvQAtRRJ34X9q4X5QjzMZgCvHwoHrj2yg2VbUnuSnmv\npJojiwZI0Mj/7y9gPGa6aBl10chs9d7V5Kaqm2qSGd/1Zf6ndSiK8P5YubP4Bvc9\nFtPgO87rbrjMmg84Trrm4EnMakQodr8GIFcQA1XqgYTrTrhlgF80/zqisN+7p80u\nfftd34IEBCOsP8ANPpxOaqOCSU+G6NbW2K6l8NYj2jhrF/cXgze25X7mMXyWj3jf\nkF0yHX8PScts1DVlx5064rrGoVQfD147m01SjWa+uqE2tD3mr9mc5QnvNdZph1CO\n3ca2jOhC2wfvqSh8cwBtGMLc1nU7K8p9qmtn3+bfL4oedwfMvyLHsa4Ggrg=\n-----END CERTIFICATE-----\n"
  MTLSKey = "-----BEGIN RSA PRIVATE KEY-----\n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
  -----END RSA PRIVATE KEY-----"
```
