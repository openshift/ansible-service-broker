# APB Filtering

APBs can be filtered out by their image name using a combination of
`white_list` or `black_list`, set on a registry basis inside the
[broker's config](config.md).

Both are optional lists of regular expressions that will be run over the
total set of discovered APBs for a given registry, determining matches.

## Filter Behavior

|     present    |                     allowed                     |                blocked               |
|:--------------:|:-----------------------------------------------:|:------------------------------------:|
| only whitelist | matches a regex in list                         | *ANY* APB that does not match        |
| only blacklist | *ALL* APBs that do not match                    | APBs that match a regex in list      |
|  both present  | matches regex in whitelist but NOT in blacklist | APBs that match a regex in blacklist |
|  None | *No* APBs from the registry | *All* APBs from that registry |

### Examples

#### Whitelist Only

```yaml
white_list:
  - "totally-legitimate.*-apb$"
  - "^my-favorite-apb$"
```

Anything matching on `totally-legitimate.*-apb$` and only `my-favorite-apb` will
be allowed through in this case. All other APBs will be **rejected**.

#### Blacklist Only

```yaml
black_list:
  - "definitely-not-malicious.*-apb$"
  - "^evil-apb$"
```

Anything matching on `definitely-not-malicious.*-apb$`and only `evil-apb` will
be blocked in this case. All other APBs will be **allowed through**.

#### Whitelist and Blacklist

```yaml
white_list:
  - "totally-legitimate.*-apb$"
  - "^my-favorite-apb$"
black_list:
  - "^totally-legitimate-rootkit-apb$"
```

Here, `totally-legitimate-rootkit-apb` is specifically blocked by the blacklist
despite its match in the whitelist because the whitelist match is overridden.

Otherwise, only those matching on `totally-legitimate.*-apb$` and
`my-favorite-apb` will be allowed through.

## Example Config File

```yaml
---
registry:
  - type: dockerhub
    name: dockerhub
    url: https://registry.hub.docker.com
    user: eriknelson
    pass: foobar
    org: eriknelson
    white_list:
      - "totally-legitimate.*-apb$"
      - "^my-favorite-apb$"
    black_list:
      - "definitely-not-malicious.*-apb$"
      - "^evil-apb$"
# ... Snipping the rest of the config file...
```
