# APB Filtering

APBs can be filtered out by their image name using a combination of
`white_list` or `black_list`, set on a registry basis inside the
[broker's config](config.md).

Both are optional lists of regular expressions that will be run over the
total set of discovered APBs for a given registry, determining matches.

## Filter Behavior

|     present    |                     allowed                     |                blocked               |
|:--------------:|:-----------------------------------------------:|:------------------------------------:|
| only allowlist | matches a regex in list                         | *ANY* APB that does not match        |
| only denylist  | *No* APBs from the registry                     | *All* APBs from that registry        |
|  both present  | matches regex in allowlist but NOT in denylist  | APBs that match a regex in denylist  |
|  None | *No* APBs from the registry | *All* APBs from that registry |

### Examples

#### Allowlist Only

```yaml
white_list:
  - "totally-legitimate.*-apb$"
  - "^my-favorite-apb$"
```

Anything matching on `totally-legitimate.*-apb$` and only `my-favorite-apb` will
be allowed through in this case. All other APBs will be **rejected**.

#### Denylist Only

```yaml
black_list:
  - "definitely-not-malicious.*-apb$"
  - "^evil-apb$"
```

Anything matching on `definitely-not-malicious.*-apb$`and only `evil-apb` will
be blocked in this case. All other APBs will be **allowed through**.

#### Allowlist and Denylist

```yaml
white_list:
  - "totally-legitimate.*-apb$"
  - "^my-favorite-apb$"
black_list:
  - "^totally-legitimate-rootkit-apb$"
```

Here, `totally-legitimate-rootkit-apb` is specifically blocked by the denylist
despite its match in the allowlist because the allowlist match is overridden.

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
