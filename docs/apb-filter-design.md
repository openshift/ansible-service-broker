# APB Filter Design Doc

## High Level Summary

Introduce 2 new fields to broker section of config:

`whitelist: "/etc/ansible-service-broker/whitelist.yaml"`
`blacklist: "/etc/ansible-service-broker/blacklist.yaml"`

Both are optional file paths that are yaml files containing arrays
of regexes. Broker will construct the set of eligible APBs from this list.

If only whitelist is present, strictest mode of operation. Any APB that does *not*
match on a regex in the whitelist will be filtered.

If only blacklist is present, all discovered APBs will be eligible *unless*
they match on a regex found in the blacklist.

If both are present, same behavior as whitelist, but if a value matches on white
and black, blacklist will take precedence and block the whitelist match.

These can be easily configured by cluster operators via config maps.

NOTE: Broker filtering of apbs should be performed behind LoadSpecs() method
on Registries. Can extend LoadSpecs() method to accept a "filter" object,
so they can easily filter discovered apbs. Registries should filter from
total list of discovered apb names before loading specs from metadata. Cuts
down on unecessary work.

Loudly log strange corner cases like no specs available due to empty but
configured files, etc.

### Cluster Admin Notes

Would normally configure these files via config maps. Can be done with separate
config maps for white and black lists in the broker. Will update templates with
examples / test scenarios for each.

## Example file:

```
# blacklist.yaml
---
- "^.*malicious.*-apb$" # Filters definitely-not-malicious-apb, malicious-apb etc.
- "bad-stuff-apb"
```

## Usage

ApbFilter struct + NewApbFilter(whitelist, blacklist).

pseudo:

```golang
// Broker setup
filter := NewApbFilter(whitelist, blacklist)
for registry := range registries {
  specs := apend(specs, registry.LoadSpecs(filter))
}

// FooRegistry
func LoadSpecs(*filter ApbFilter) ([]*Spec, int, error) {
  allNames := discoverApbNames()
  validNames, filteredNames, err = filter.FilterApbNames(allNames)
  log_filtered(filteredApbs)
  log_valid(validNames)
  return loadSpecs(validNames)
}
```

Could make this some kind of shared behavior registry's get for free.

# Questions

* Multi-registry considerations. Should you be able to black/whitelist
by registry? I.E., whitelist private_registry/foo-apb, blacklist docker.io/foo-apb

# Stretch goals
* Extend broker API to dump an unfiltered list of discovered apbs to assist
cluster ops in creating white/blacklists
