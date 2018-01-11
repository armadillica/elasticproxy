# ElasticProxy Changelog

## Version 1.2 (2018-01-11)

- Updated allowed methods to be compatible with Kibana 6.1


## Version 1.1 (2017-12-29)

- Added whitelist for GET requests based on path prefixes. This allows us to run Kibana with Cloud
  statistics as usual, but disallows any access to other indices (such as users and nodes).


## Version 1.0 (2017-10-03)

- First publicly used version.
