Subject: support pagination in the /catalog API

Filed: https://github.com/openservicebrokerapi/servicebroker/issues/126

# Proposal
The catalog API should support pagination in the cases where there are large
numbers of services supported by a broker. We see a potential for 10k-15k
catalog entries which could take a bit of time to render to the caller.

Effectively the initial catalog call should expect the Link header or the
metadata in the body indicating that the broker should be invoked with
pagination.

The API should take the following optional parameters:

parameter name | description
-------------- | -----------
per_page | number of items to show per page
page | page number requested, depends on page size. if > last page, then last page is returned

### cURL
```
   $ curl -H "X-Broker-API-Version: 2.9" http://username:password@broker-url/v2/catalog?page=2&per_page=20
```

## Body (changes in bold)
Response field | Type | Description
-------------- | ---- | ------------
services* | array-of-service-objects | Schema of service objects defined below.
**pagination** | **metadata about pagination** | Schema for pagination defined below.

### Response
The ```pagination``` block would be added to the catalog response at the same
level as ```services```. In the case of requesting the first page, the response
will contain a *next* and a *last* item.

```
{
  "services": [{
     ...
     }],
  "pagination": {
    "next": "https://broker/v2/catalog?page=2",
    "last": "https://broker/v2/catalog?page=15333"
  }
}
```

Alternatively we could use the Link header [RFC 5988](https://tools.ietf.org/html/rfc5988). In this case there would be **NO** changes to the response body just the headers.

```
Link: <https://broker/v2/catalog?page=2>; rel="next",
      <https://broker/v2/catalog?page=15333>; rel="last"
```

Subsequent calls will yield 2 new options, *first* and *prev*. For example, if you request page 10000 the following response will be returned:

```
{
  "services": [{
     ...
     }],
  "pagination": {
    "next": "https://broker/v2/catalog?page=10001",
    "last": "https://broker/v2/catalog?page=15333",
    "first": "https://broker/v2/catalog?page=1",
    "prev":  "https://broker/v2/catalog?page=9999"
  }
}
```
Alternatively we could use the Link header [RFC 5988](https://tools.ietf.org/html/rfc5988)

```
Link: <https://broker/v2/catalog?page=10001>; rel="next",
      <https://broker/v2/catalog?page=15333>; rel="last",
      <https://broker/v2/catalog?page=1>; rel="first",
      <https://broker/v2/catalog?page=9999>; rel="prev"
```
