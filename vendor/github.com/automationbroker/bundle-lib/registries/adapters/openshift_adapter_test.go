//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package adapters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	ft "github.com/stretchr/testify/assert"
)

var Images = []string{"satoshi/nakamoto", "foo/bar", "paul/atreides"}

const OpenShiftManifestResponse = `
{
   "schemaVersion": 1,
   "name": "%v",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
      {
         "blobSum": "sha256:74d70fd19a822808f93dac84e4ebe178883cf03b2be3f4e1957070d8a8d4505f"
      },
      {
         "blobSum": "sha256:86c2e2710c6869f55e4c4852d2a4416f50c38df8b538750fa83037090b8f1a5e"
      },
      {
         "blobSum": "sha256:0001a3087112018853b83f67ffc311dab755d14393a69852d5e2f4aa01b35361"
      },
      {
         "blobSum": "sha256:4e5a7647df476dcb309aa02f6901239300e7103a914fd92acf540372c1dafe0c"
      }
   ],
   "history": [
      {
         "v1Compatibility": "{\"architecture\":\"amd64\",\"author\":\"Ansible Playbook Bundle Community\",\"config\":{\"Hostname\":\"82021aceee3e\",\"Domainname\":\"\",\"User\":\"apb\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"container=oci\",\"USER_NAME=apb\",\"USER_UID=1001\",\"BASE_DIR=/opt/apb\",\"HOME=/opt/apb\"],\"Cmd\":null,\"ArgsEscaped\":true,\"Image\":\"a3614abb513481e0d2ce1915cfc798b0655f97370ec73ac4d51c09ff4260775e\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":[\"entrypoint.sh\"],\"OnBuild\":[],\"Labels\":{\"architecture\":\"x86_64\",\"authoritative-source-url\":\"registry.access.redhat.com\",\"build-date\":\"2017-06-16T17:13:27.381723\",\"com.redhat.apb.spec\":\"aWQ6IGUxYmNkNGE4LWNlMDItNDU4NS05ZjRjLTE4YWJkNTZkNzZmMgpuYW1lOiBwb3N0Z3Jlc3FsLWFwYgppbWFnZTogb3BlbnNoaWZ0My9wb3N0Z3Jlc3FsLWFwYgpkZXNjcmlwdGlvbjogU0NMIFBvc3RncmVTUUwgYXBiIGltcGxlbWVudGF0aW9uCmJpbmRhYmxlOiB0cnVlCmFzeW5jOiBvcHRpb25hbAptZXRhZGF0YToKICBkaXNwbGF5TmFtZTogIlBvc3RncmVTUUwgKEFQQikiCiAgbG9uZ0Rlc2NyaXB0aW9uOiAiQW4gYXBiIHRoYXQgZGVwbG95cyBwb3N0Z3Jlc3FsIDkuNCBvciA5LjUuIgogIGNvbnNvbGUub3BlbnNoaWZ0LmlvL2ljb25DbGFzczogaWNvbi1wb3N0Z3Jlc3FsCiAgZG9jdW1lbnRhdGlvblVybDogImh0dHBzOi8vd3d3LnBvc3RncmVzcWwub3JnL2RvY3MvIgp0YWdzOgogIC0gZGF0YWJhc2VzCiAgLSBwb3N0Z3Jlc3FsCnBhcmFtZXRlcnM6CiAgLSBwb3N0Z3Jlc3FsX2RhdGFiYXNlOgogICAgICB0aXRsZTogUG9zdGdyZVNRTCBEYXRhYmFzZSBOYW1lCiAgICAgIHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBhZG1pbgogIC0gcG9zdGdyZXNxbF9wYXNzd29yZDoKICAgICAgdGl0bGU6IFBvc3RncmVTUUwgUGFzc3dvcmQKICAgICAgZGVzY3JpcHRpb246IEEgcmFuZG9tIGFscGhhbnVtZXJpYyBzdHJpbmcgaWYgbGVmdCBibGFuawogICAgICB0eXBlOiBzdHJpbmcKICAgICAgZGVmYXVsdDogYWRtaW4KICAtIHBvc3RncmVzcWxfdXNlcjoKICAgICAgdGl0bGU6IFBvc3RncmVTUUwgVXNlcgogICAgICB0eXBlOiBzdHJpbmcKICAgICAgZGVmYXVsdDogYWRtaW4KICAgICAgbWF4bGVuZ3RoOiA2MwogIC0gcG9zdGdyZXNxbF92ZXJzaW9uOgogICAgICB0aXRsZTogUG9zdGdyZVNRTCBWZXJzaW9uCiAgICAgIHR5cGU6IGVudW0KICAgICAgZGVmYXVsdDogOS41CiAgICAgIGVudW06IFsiOS41IiwgIjkuNCJdCnJlcXVpcmVkOgogIC0gcG9zdGdyZXNxbF9kYXRhYmFzZQogIC0gcG9zdGdyZXNxbF91c2VyCiAgLSBwb3N0Z3Jlc3FsX3ZlcnNpb24K\",\"com.redhat.apb.version\":\"0.1.0\",\"com.redhat.build-host\":\"rcm-img-docker02.build.eng.bos.redhat.com\",\"com.redhat.component\":\"openshift-enterprise-postgresql-apb\",\"description\":\"The Red Hat Enterprise Linux Base image is designed to be a fully supported foundation for your containerized applications.  This base image provides your operations and application teams with the packages, language runtimes and tools necessary to run, maintain, and troubleshoot all of your applications. This image is maintained by Red Hat and updated regularly. It is designed and engineered to be the base layer for all of your containerized applications, middleware and utilites. When used as the source for all of your containers, only one copy will ever be downloaded and cached in your production environment. Use this image just like you would a regular Red Hat Enterprise Linux distribution. Tools like yum, gzip, and bash are provided by default. For further information on how this image was built look at the /root/anacanda-ks.cfg file.\",\"distribution-scope\":\"public\",\"io.k8s.display-name\":\"Red Hat Enterprise Linux 7\",\"io.openshift.tags\":\"base rhel7\",\"name\":\"openshift3-tech-preview/openshift-enterprise-postgresql-apb\",\"release\":\"5\",\"summary\":\"Provides the latest release of Red Hat Enterprise Linux 7 in a fully featured and supported base image.\",\"vcs-ref\":\"3c7cfc8a7a1fc3e1da20ed4460e3bb70dd93c67d\",\"vcs-type\":\"git\",\"vendor\":\"Red Hat, Inc.\",\"version\":\"0.0.1\"}},\"container_config\":{\"Hostname\":\"82021aceee3e\",\"Domainname\":\"\",\"User\":\"apb\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"container=oci\",\"USER_NAME=apb\",\"USER_UID=1001\",\"BASE_DIR=/opt/apb\",\"HOME=/opt/apb\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) USER [apb]\"],\"ArgsEscaped\":true,\"Image\":\"sha256:edfa21ad402393d6a9aed7961306908f9eab3f23c1e8fefa2f6ab4d80238b31b\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":[\"entrypoint.sh\"],\"OnBuild\":[],\"Labels\":{\"architecture\":\"x86_64\",\"authoritative-source-url\":\"registry.access.redhat.com\",\"build-date\":\"2017-06-16T17:13:27.381723\",\"com.redhat.apb.spec\":\"aWQ6IGUxYmNkNGE4LWNlMDItNDU4NS05ZjRjLTE4YWJkNTZkNzZmMgpuYW1lOiBwb3N0Z3Jlc3FsLWFwYgppbWFnZTogb3BlbnNoaWZ0My9wb3N0Z3Jlc3FsLWFwYgpkZXNjcmlwdGlvbjogU0NMIFBvc3RncmVTUUwgYXBiIGltcGxlbWVudGF0aW9uCmJpbmRhYmxlOiB0cnVlCmFzeW5jOiBvcHRpb25hbAptZXRhZGF0YToKICBkaXNwbGF5TmFtZTogIlBvc3RncmVTUUwgKEFQQikiCiAgbG9uZ0Rlc2NyaXB0aW9uOiAiQW4gYXBiIHRoYXQgZGVwbG95cyBwb3N0Z3Jlc3FsIDkuNCBvciA5LjUuIgogIGNvbnNvbGUub3BlbnNoaWZ0LmlvL2ljb25DbGFzczogaWNvbi1wb3N0Z3Jlc3FsCiAgZG9jdW1lbnRhdGlvblVybDogImh0dHBzOi8vd3d3LnBvc3RncmVzcWwub3JnL2RvY3MvIgp0YWdzOgogIC0gZGF0YWJhc2VzCiAgLSBwb3N0Z3Jlc3FsCnBhcmFtZXRlcnM6CiAgLSBwb3N0Z3Jlc3FsX2RhdGFiYXNlOgogICAgICB0aXRsZTogUG9zdGdyZVNRTCBEYXRhYmFzZSBOYW1lCiAgICAgIHR5cGU6IHN0cmluZwogICAgICBkZWZhdWx0OiBhZG1pbgogIC0gcG9zdGdyZXNxbF9wYXNzd29yZDoKICAgICAgdGl0bGU6IFBvc3RncmVTUUwgUGFzc3dvcmQKICAgICAgZGVzY3JpcHRpb246IEEgcmFuZG9tIGFscGhhbnVtZXJpYyBzdHJpbmcgaWYgbGVmdCBibGFuawogICAgICB0eXBlOiBzdHJpbmcKICAgICAgZGVmYXVsdDogYWRtaW4KICAtIHBvc3RncmVzcWxfdXNlcjoKICAgICAgdGl0bGU6IFBvc3RncmVTUUwgVXNlcgogICAgICB0eXBlOiBzdHJpbmcKICAgICAgZGVmYXVsdDogYWRtaW4KICAgICAgbWF4bGVuZ3RoOiA2MwogIC0gcG9zdGdyZXNxbF92ZXJzaW9uOgogICAgICB0aXRsZTogUG9zdGdyZVNRTCBWZXJzaW9uCiAgICAgIHR5cGU6IGVudW0KICAgICAgZGVmYXVsdDogOS41CiAgICAgIGVudW06IFsiOS41IiwgIjkuNCJdCnJlcXVpcmVkOgogIC0gcG9zdGdyZXNxbF9kYXRhYmFzZQogIC0gcG9zdGdyZXNxbF91c2VyCiAgLSBwb3N0Z3Jlc3FsX3ZlcnNpb24K\",\"com.redhat.apb.version\":\"0.1.0\",\"com.redhat.build-host\":\"rcm-img-docker02.build.eng.bos.redhat.com\",\"com.redhat.component\":\"openshift-enterprise-postgresql-apb\",\"description\":\"The Red Hat Enterprise Linux Base image is designed to be a fully supported foundation for your containerized applications.  This base image provides your operations and application teams with the packages, language runtimes and tools necessary to run, maintain, and troubleshoot all of your applications. This image is maintained by Red Hat and updated regularly. It is designed and engineered to be the base layer for all of your containerized applications, middleware and utilites. When used as the source for all of your containers, only one copy will ever be downloaded and cached in your production environment. Use this image just like you would a regular Red Hat Enterprise Linux distribution. Tools like yum, gzip, and bash are provided by default. For further information on how this image was built look at the /root/anacanda-ks.cfg file.\",\"distribution-scope\":\"public\",\"io.k8s.display-name\":\"Red Hat Enterprise Linux 7\",\"io.openshift.tags\":\"base rhel7\",\"name\":\"openshift3-tech-preview/openshift-enterprise-postgresql-apb\",\"release\":\"5\",\"summary\":\"Provides the latest release of Red Hat Enterprise Linux 7 in a fully featured and supported base image.\",\"vcs-ref\":\"3c7cfc8a7a1fc3e1da20ed4460e3bb70dd93c67d\",\"vcs-type\":\"git\",\"vendor\":\"Red Hat, Inc.\",\"version\":\"0.0.1\"}},\"created\":\"2017-06-16T17:25:15.720997Z\",\"docker_version\":\"1.10.3\",\"id\":\"c839b2d43aeee633b52d61b40792db20f97b4e5eef9e3ec923d2d97fe94a97d4\",\"os\":\"linux\",\"parent\":\"9e14c200af07a3ba7262290ccb6963709631fb65ed76ea4bb679c1249e089be1\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"9e14c200af07a3ba7262290ccb6963709631fb65ed76ea4bb679c1249e089be1\",\"parent\":\"d70bfda7a2e2e81af2046621974d6cfbafb623da3027323968ebfaed378f4d81\",\"created\":\"2017-06-16T16:18:11.58354Z\",\"container_config\":{\"Cmd\":[\"/bin/sh -c rm -f '/etc/yum.repos.d/asb-apb-unsigned-ose-3.6.repo'\"]},\"author\":\"Ansible Playbook Bundle Community\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"d70bfda7a2e2e81af2046621974d6cfbafb623da3027323968ebfaed378f4d81\",\"parent\":\"fc1d9e6b0ae5bcc28707acbd27a80f73bef6c9a9ebc665608b13fc2069d9b9d3\",\"created\":\"2017-05-18T16:00:41.296037Z\",\"container_config\":{\"Cmd\":[\"/bin/sh -c rm -f '/etc/yum.repos.d/compose-rpms-1.repo'\"]},\"author\":\"Red Hat, Inc.\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"fc1d9e6b0ae5bcc28707acbd27a80f73bef6c9a9ebc665608b13fc2069d9b9d3\",\"comment\":\"Imported from -\",\"created\":\"2017-05-18T15:59:20.383772669Z\",\"container_config\":{\"Cmd\":[\"\"]}}"
      }
   ],
   "signatures": [
      {
         "header": {
            "jwk": {
               "crv": "P-256",
               "kid": "4YQB:KEUP:4MSX:HAD7:BADG:LC4F:5RFH:EQMC:ZLKI:XCAP:WGCJ:SDCB",
               "kty": "EC",
               "x": "jdv0lbVXbFOwP-PR3jgzHi0VITq9uf_P5aKTyYBNGTY",
               "y": "HcPE_Gm8QAvAL_ULuC1L-_FRBODua_Rn2pQgAjyjQ8g"
            },
            "alg": "ES256"
         },
         "signature": "4DuXenHioqVKeqdaqqYHmflj1zJu0ZSlRbfYv8xlFahVP_ZllFBpDjU8CX4DHNGov4BoEuXLfRqqbFlpN4NzFw",
         "protected": "eyJmb3JtYXRMZW5ndGgiOjkyMzQsImZvcm1hdFRhaWwiOiJDbjAiLCJ0aW1lIjoiMjAxNy0wNi0yMFQwNTowNzoxOVoifQ"
      }
   ]
 }`

const AuthResponse = `
{
  "token": "ertssGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwiandrIjogeyJrdHkiOiJSU0EiLCJhbGciOiJSUzI1NiIsInVzZSI6InNpZyIsIm4iOiI1TWhlWnh2VU9id0NJRXVYZWJubjlNQ0JFcDlTdWUyME44SHhQTExSWjdYNUl1QTU0a3dpVUQtMXBqZ2tpamgwRXpTZGlxbTdWUExlNGl1eXZEZmJxX1d2QWNNUjVhb3RKbl9EMGllbnA5eVRpR3d5SEVQWG83MERwZ2FsN1ZOc0RCdWdTSF91SThpM3BVaHZLczAxMEwzRVFmRVkwc2hrQVZRY1hEb3JNVmhkZkd2WE84aURMODBqR1BDWmNidXpZc1BQQ0RycWd5UTl6SFctWGVMSURXUlh3YW16TE1pNGZqOXRNcXB6b3FUdkhZUUFGQWdHbTU2Umh0ZDE0SWxaU19yYnZ2aFFLMDVEUDB2MHJtZFU0SXZoTzBTcUJLRjdkN1FIbFppd09nd2ppQzktU3JORkFzSngtSy1PMzBROThpMUhKRjd2MXNCMFE1OTFzYUVZMmM3VWhXTnpqdm5DM2gtYXdtSi1TM2FJSFFGNWNObmtTbTRUN3VZWHEwRW9oMEM2cEdlMUlZSzM1RHlGZjRsZEZXVTNVQmpvUlc0dG1jd2lmNE96ZXg3eWw3NmV5WWRxeXRscXk0bmwxeWZlTXpMSTJlNm5ucWlBdFdEUTY2Y3NSTVZVb1JpdUtvZWdOZldIelRUdUZabU1OdHpNdXpsR1JFbmtJazZnZl84VUNPME9Kd1FZM2luQldGZFc2V2l4RWRwSzFxT1VEVzBjWGRQVWt0NkZhNHJzU0NUTTZvbWNvVkRiU3hWV3Fqa2JtS0ZnU1pnVjdaUUZiM2doa3E5VmhtZVlzbmpfSmc0Skxnc3JpRzlZUWlKOWljYl9TWHNuNUhJQXcyOWhFR3BEMW01Q05IZmt6SVBxbllYWmlfaXAwSGlwWjBiVU9RN3JBYXVkTHktZHlhcyIsImUiOiJBUUFCIn19.eyJqdGkiOiJlZjgxMTgxOS04ZTE3LTRjZTQtYWY2OC0yY2E3M2ZlYzc2N2IiLCJleHAiOjE1MjIxODA3NTksIm5iZiI6MTUyMjE4MDQ1OSwiaWF0IjoxNTIyMTgwNDU5LCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JoYzR0cCIsImF1ZCI6ImRvY2tlci1yZWdpc3RyeSIsInN1YiI6IlF1YWxpdHlBc3N1cmFuY2UiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJkb2NrZXItcmVnaXN0cnkiLCJhY2Nlc3MiOltdfQ.kI0kkiUP-M6epKa9G7pm9-8fUb9oM2aP93bObL5SackVO1pv0dlk6Kqml1X5W3pgV_tabRDRA_Oc5rdCImOXdCz1J8SnSO3wNeX02JgWLC4LWRUnJuNQtqIHdYSD3YdmOcsou5psD2I_gv-QhsaJtb5frCklJ5gU68i0SsUhPr0sf1BDD1oVzHu_US-zNuwXbzxiug4C5oObTLn9vzfz8oE4Hwy-WIRC0LeJxdNIZvyaUyiBEA7fq3p3XlInvsmgv94l-HZiEtOi7Gml_EfDHj4G7pZGNOgfoTHhR8ugiE9rVTEGo9esUNaRW7XhGCQV2CH3IowmtrIAMndToI4yNLPpOi8alCsaomuIunyGDhGqmdT2iKBrFErliN9b7oof73ElqJdjZhWD5_8iWt7ICQFAfomVe_TU8UdXlp5mYRf2VlF-IMk1YbUZh8aMFIJ1lZsmCzM9sZEpv7Y34AI837nDwL1vwWHVaFVLrW3oLtIXVmnvbLYwnNw7Gus6Gxh5mQI_ZGcSXSmOcK-mGFfVJgsJPTVhJI4FAQMsxdJwL8yAM96FCN2h9jcgfk0eQ1T4BRQFNLsy51dH1aUPYW51xR2GnbU6aahwjjYX3CQs6akPsNCnvQc6WaRb0IUwFrkmGBuIennb8W1A2opNRetVk3r1kiTm4zNaYrsF67PwTqg",
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwiandrIjogeyJrdHkiOiJSU0EiLCJhbGciOiJSUzI1NiIsInVzZSI6InNpZyIsIm4iOiI1TWhlWnh2VU9id0NJRXVYZWJubjlNQ0JFcDlTdWUyME44SHhQTExSWjdYNUl1QTU0a3dpVUQtMXBqZ2tpamgwRXpTZGlxbTdWUExlNGl1eXZEZmJxX1d2QWNNUjVhb3RKbl9EMGllbnA5eVRpR3d5SEVQWG83MERwZ2FsN1ZOc0RCdWdTSF91SThpM3BVaHZLczAxMEwzRVFmRVkwc2hrQVZRY1hEb3JNVmhkZkd2WE84aURMODBqR1BDWmNidXpZc1BQQ0RycWd5UTl6SFctWGVMSURXUlh3YW16TE1pNGZqOXRNcXB6b3FUdkhZUUFGQWdHbTU2Umh0ZDE0SWxaU19yYnZ2aFFLMDVEUDB2MHJtZFU0SXZoTzBTcUJLRjdkN1FIbFppd09nd2ppQzktU3JORkFzSngtSy1PMzBROThpMUhKRjd2MXNCMFE1OTFzYUVZMmM3VWhXTnpqdm5DM2gtYXdtSi1TM2FJSFFGNWNObmtTbTRUN3VZWHEwRW9oMEM2cEdlMUlZSzM1RHlGZjRsZEZXVTNVQmpvUlc0dG1jd2lmNE96ZXg3eWw3NmV5WWRxeXRscXk0bmwxeWZlTXpMSTJlNm5ucWlBdFdEUTY2Y3NSTVZVb1JpdUtvZWdOZldIelRUdUZabU1OdHpNdXpsR1JFbmtJazZnZl84VUNPME9Kd1FZM2luQldGZFc2V2l4RWRwSzFxT1VEVzBjWGRQVWt0NkZhNHJzU0NUTTZvbWNvVkRiU3hWV3Fqa2JtS0ZnU1pnVjdaUUZiM2doa3E5VmhtZVlzbmpfSmc0Skxnc3JpRzlZUWlKOWljYl9TWHNuNUhJQXcyOWhFR3BEMW01Q05IZmt6SVBxbllYWmlfaXAwSGlwWjBiVU9RN3JBYXVkTHktZHlhcyIsImUiOiJBUUFCIn19.eyJqdGkiOiJlZjgxMTgxOS04ZTE3LTRjZTQtYWY2OC0yY2E3M2ZlYzc2N2IiLCJleHAiOjE1MjIxODA3NTksIm5iZiI6MTUyMjE4MDQ1OSwiaWF0IjoxNTIyMTgwNDU5LCJpc3MiOiJodHRwczovL3Nzby5yZWRoYXQuY29tL2F1dGgvcmVhbG1zL3JoYzR0cCIsImF1ZCI6ImRvY2tlci1yZWdpc3RyeSIsInN1YiI6IlF1YWxpdHlBc3N1cmFuY2UiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJkb2NrZXItcmVnaXN0cnkiLCJhY2Nlc3MiOltdfQ.kI0kkiUP-M6epKa9G7pm9-8fUb9oM2aP93bObL5SackVO1pv0dlk6Kqml1X5W3pgV_tabRDRA_Oc5rdCImOXdCz1J8SnSO3wNeX02JgWLC4LWRUnJuNQtqIHdYSD3YdmOcsou5psD2I_gv-QhsaJtb5frCklJ5gU68i0SsUhPr0sf1BDD1oVzHu_US-zNuwXbzxiug4C5oObTLn9vzfz8oE4Hwy-WIRC0LeJxdNIZvyaUyiBEA7fq3p3XlInvsmgv94l-HZiEtOi7Gml_EfDHj4G7pZGNOgfoTHhR8ugiE9rVTEGo9esUNaRW7XhGCQV2CH3IowmtrIAMndToI4yNLPpOi8alCsaomuIunyGDhGqmdT2iKBrFErliN9b7oof73ElqJdjZhWD5_8iWt7ICQFAfomVe_TU8UdXlp5mYRf2VlF-IMk1YbUZh8aMFIJ1lZsmCzM9sZEpv7Y34AI837nDwL1vwWHVaFVLrW3oLtIXVmnvbLYwnNw7Gus6Gxh5mQI_ZGcSXSmOcK-mGFfVJgsJPTVhJI4FAQMsxdJwL8yAM96FCN2h9jcgfk0eQ1T4BRQFNLsy51dH1aUPYW51xR2GnbU6aahwjjYX3CQs6akPsNCnvQc6WaRb0IUwFrkmGBuIennb8W1A2opNRetVk3r1kiTm4zNaYrsF67PwTqg",
  "expires_in": 300,
  "issued_at": "2018-03-27T19:54:19Z"
}`

func TestRegistryName(t *testing.T) {
	testcases := []struct {
		urlstr string
		result string
	}{
		{"http://foo.redhat.com", "foo.redhat.com"},
		{"https://registry.reg-aws.openshift.com:443", "registry.reg-aws.openshift.com:443"},
		{"ftp://registry.reg-aws.openshift.com", "registry.reg-aws.openshift.com"},
		{"file:///home/user/awesome", "/home/user/awesome"},
	}

	for _, testcase := range testcases {
		url, err := url.Parse(testcase.urlstr)
		if err != nil {
			t.Fatal("Error: ", err)
		}

		ocpa := OpenShiftAdapter{
			Config: Configuration{
				URL: url,
			},
		}
		ft.Equal(t, ocpa.RegistryName(), testcase.result, "the registry name returned did not match expected value")
	}
}

func TestOpenShiftGetImageNames(t *testing.T) {
	ocpa := OpenShiftAdapter{}
	ocpa.Config.Images = Images
	imagesFound, err := ocpa.GetImageNames()
	if err != nil {
		t.Fatal("Error: ", err)
	}
	ft.Equal(t, imagesFound, Images, "image names returned did not match expected config")
}

func TestOpenShiftFetchSpecs(t *testing.T) {
	authServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected `GET` request, got `%s`", r.Method)
		}
		fmt.Fprintf(w, AuthResponse)
	}))

	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected `GET` request, got `%s`", r.Method)
		}

		if strings.HasSuffix(r.URL.EscapedPath(), "/v2/") {
			w.Header().Add("Www-Authenticate", fmt.Sprintf("Bearer realm=\"%v/v2/auth/realms/foo-docker-v2/auth\",service=\"docker-registry\"", authServ.URL))
		}
		if strings.Contains(r.URL.EscapedPath(), "manifests/") {
			name := strings.Split(r.URL.EscapedPath(), "manifests/")[1]
			fmt.Fprintf(w, fmt.Sprintf(OpenShiftManifestResponse, name))
		}

	}))

	url, err := url.Parse(serv.URL)
	if err != nil {
		t.Fatal("Error: ", err)
	}
	ocpa := OpenShiftAdapter{
		Config: Configuration{
			URL:    url,
			User:   "satoshi",
			Pass:   "nakamoto",
			Images: Images,
		},
	}
	imageNames, err := ocpa.GetImageNames()
	specs, err := ocpa.FetchSpecs(imageNames)
	if err != nil {
		t.Fatal("Error: ", err)
	}
	if specs == nil {
		t.Fatal("Error: did not find fetch any valid specs")
	}
	if len(specs) != 3 {
		t.Fatal("Error: did not find 3 expected specs, only found: ", len(specs))
	}
}
