<!-- Your Title -->
<p align="center">
<a href="https://www.selefra.io/" target="_blank">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/124020340/225567784-61adb5e7-06ae-402a-9907-69c1e6f1aa9e.png"><source media="(prefers-color-scheme: light)" srcset="https://user-images.githubusercontent.com/124020340/224677116-44ae9c6c-a543-4813-9ef3-c7cbcacd2fbe.png"><img width="400px" alt="Steampipe Logo" src="https://user-images.githubusercontent.com/124020340/224677116-44ae9c6c-a543-4813-9ef3-c7cbcacd2fbe.png"></picture>
<a/>
</p>

<!-- Description -->
  <p align="center">
    <i>Selefra is an open-source policy-as-code software that provides analytics for multi-cloud and SaaS.</i>
  </p>

  <!-- Badges -->
<p align="center">   
<a href="https://pkg.go.dev/github.com/selefra/selefra"><img alt="go" src="https://img.shields.io/badge/go-1.19-1E90FF" /></a>
<a href="https://github.com/selefra/selefra"><img alt="Total" src="https://img.shields.io/github/downloads/selefra/selefra/total?logo=github"></a>
<a href="https://github.com/selefra/selefra/blob/master/LICENSE"><img alt="GitHub license" src="https://img.shields.io/github/license/selefra/selefra?style=social"></a>
  </p>

  <!-- Badges -->
  <p align="center">
<a href="https://selefra.io/community/join"><img src="https://img.shields.io/badge/-Slack-424549?style=social&logo=Slack" height=25></a>
    &nbsp;
    <a href="https://twitter.com/SelefraCorp"><img src="https://img.shields.io/badge/-Twitter-red?style=social&logo=twitter" height=25></a>
    &nbsp;
    <a href="https://www.reddit.com/r/Selefra"><img src="https://img.shields.io/badge/-Reddit-red?style=social&logo=reddit" height=25></a>
    &nbsp;
    <a href="https://selefra.medium.com/"><img src="https://img.shields.io/badge/-Medium-red?style=social&logo=medium" height=25></a>

  </p>

<p align="center">
  <img width="900" alt="banner" src="https://user-images.githubusercontent.com/124020340/232656647-58e2c31f-ba94-48f0-99fc-87ab660309d0.png">
</p>
<br/>

<!-- About Selefra -->

## About Selefra

Selefra means "select * from infrastructure". It is an open-source policy-as-code software that provides analysis for multi-cloud and SaaS environments, including over 30 services such as AWS, GCP, Azure, Alibaba Cloud, Kubernetes, Github, Cloudflare, and Slack.

For best practices and detailed instructions, refer to the Docs. Within the [Docs](https://selefra.io/docs/introduction), you will find information on installation, CLI usage, project workflow, and more guides on how to accomplish cloud inspection tasks.

With Selefra, you can engage in conversations with GPT models, which will analyze the information and provide relevant suggestions for security, cost, and architecture checks, helping you better manage their cloud resources, enhance security, reduce costs, and optimize architecture design.

<img align="right" width="570" alt="img_code" src="https://user-images.githubusercontent.com/124020340/232016353-67b21268-ae70-47a9-a848-cad0f2fce66f.gif">

#### 🔥 Policy as Code

Custom analysis policies (security, compliance, cost) can be written through a combination of SQL and YAML.

#### 💥 Configuration of Multi-Cloud, Multi-SaaS

Unified multi-cloud configuration data integration capabilities that can support analysis of configuration data from any cloud service via SQL.

#### 🌟 Version Control

Analysis policies can be managed through VCS such as GitHub/Gitlab.

#### 🥤 Automation

Policies can be automated to enforce compliance, security, and cost optimization rules through Scheduled tasks and cloud automation tools.

## Getting started

Read detailed documentation for how to [Get Started](https://selefra.io/docs/get-started/) with Selefra.

For quick start, run this demo, it should take less than a few minutes:

1. **Install Selefra**

   For non-macOS users, [download packages](https://github.com/selefra/selefra/releases) to install Selefra.

   On macOS, tap Selefra with Homebrew:

    ```bash
    brew tap selefra/tap
    ```

   Next, install Selefra:

    ```bash
    brew install selefra/tap/selefra
    ```

2. **Initialization project**

    ```bash
    mkdir selefra-demo && cd selefra-demo && selefra init
    ```

3. **Build code**

    ```bash
    selefra apply 
    ```
   
## 🔥 Analyze cloud resources using GPT

You can refer to the [documentation](https://selefra.io/docs/get-started#use-gpt)  to configure your OPENAPI_API_KEY in advance and start analyzing your cloud resources

```bash
selefra gpt <"what you want to analyze"> --openai_mode=gpt-3.5 --openai_limit=5 --openai_api_key=<Your Openai Api Key>
```

## Selefra Community Ecosystem

 Provider | Introduce | Status |
 | --------| ----- | ------ |
| [AWS](https://www.selefra.io/docs/providers-connector/aws)|The AWS Provider for Selefra can be used to extract data from many of the cloud services by AWS. The provider must be configured with credentials to extract and analyze infrastructure data from AWS. | Stable |
| [GCP](https://www.selefra.io/docs/providers-connector/gcp)|The GCP Provider for Selefra can be used to extract data from many of the cloud services by GCP. The provider must be configured with credentials to extract and analyze infrastructure data from GCP. | Stable |
| [K8S](https://www.selefra.io/docs/providers-connector/k8s)|The K8s Provider for Selefra can be used to extract data from many of the cloud services by K8s. The provider must be configured with credentials to extract and analyze infrastructure data from K8s. | Stable |
| [Azure](https://www.selefra.io/docs/providers-connector/azure)| The Azure Provider for Selefra can be used to extract data from many of the cloud services by Azure. The provider must be configured with credentials to extract and analyze infrastructure data from Azure.    | Stable |
| [Slack](https://www.selefra.io/docs/providers-connector/slack)| The Slack Provider for Selefra can be used to extract data from many of the cloud services by Slack. The provider must be configured with credentials to extract and analyze infrastructure data from Slack.    | Stable |
| [Cloudflare](https://www.selefra.io/docs/providers-connector/cloudflare)| The Cloudflare Provider for Selefra can be used to extract data from many of the cloud services by Cloudflare. The provider must be configured with credentials to extract and analyze infrastructure data from Cloudflare.    | Stable |
| [Datadog](https://www.selefra.io/docs/providers-connector/datadog)| The Datadog Provider for Selefra can be used to extract data from many of the cloud services by Datadog. The provider must be configured with credentials to extract and analyze infrastructure data from Datadog.    | Stable |
| [Microsoft365](https://www.selefra.io/docs/providers-connector/microsoft365)| The Microsoft365 Provider for Selefra can be used to extract data from many of the cloud services by Microsoft365. The provider must be configured with credentials to extract and analyze infrastructure data from Microsoft365.    | Stable |
| [Vercel](https://www.selefra.io/docs/providers-connector/vercel)| The Vercel Provider for Selefra can be used to extract data from many of the cloud services by Vercel. The provider must be configured with credentials to extract and analyze infrastructure data from Vercel.    | Stable |
| [Github](https://www.selefra.io/docs/providers-connector/github)| The Github Provider for Selefra can be used to extract data from many of the cloud services by Github. The provider must be configured with credentials to extract and analyze infrastructure data from Github.    | Stable |
| [GoogleWorksplace](https://www.selefra.io/docs/providers-connector/googleworksplace)| The GoogleWorksplace Provider for Selefra can be used to extract data from many of the cloud services by GoogleWorksplace. The provider must be configured with credentials to extract and analyze infrastructure data from GoogleWorksplace.    | Stable |
| [Auth0](https://www.selefra.io/docs/providers-connector/auth0)| The Auth0 Provider for Selefra can be used to extract data from many of the cloud services by Auth0. The provider must be configured with credentials to extract and analyze infrastructure data from Auth0.    | Stable |
| [Zendesk](https://www.selefra.io/docs/providers-connector/zendesk)| The Zendesk Provider for Selefra can be used to extract data from many of the cloud services by Zendesk. The provider must be configured with credentials to extract and analyze infrastructure data from Zendesk.    | Stable |
| [Consul](https://www.selefra.io/docs/providers-connector/consul)| The Consul Provider for Selefra can be used to extract data from many of the cloud services by Consul. The provider must be configured with credentials to extract and analyze infrastructure data from Consul.    | Stable |
| [Zoom](https://www.selefra.io/docs/providers-connector/zoom)| The Zoom Provider for Selefra can be used to extract data from many of the cloud services by Zoom. The provider must be configured with credentials to extract and analyze infrastructure data from Zoom.    | Stable |
| [Gandi](https://www.selefra.io/docs/providers-connector/gandi)| The Gandi Provider for Selefra can be used to extract data from many of the cloud services by Gandi. The provider must be configured with credentials to extract and analyze infrastructure data from Gandi.    | Stable |
| [Heroku](https://www.selefra.io/docs/providers-connector/heroku)| The Heroku Provider for Selefra can be used to extract data from many of the cloud services by Heroku. The provider must be configured with credentials to extract and analyze infrastructure data from Heroku.    | Stable |
| [IBM](https://www.selefra.io/docs/providers-connector/ibm)| The IBM Provider for Selefra can be used to extract data from many of the cloud services by IBM. The provider must be configured with credentials to extract and analyze infrastructure data from IBM.    | Stable |
| [Pagerduty](https://www.selefra.io/docs/providers-connector/pagerduty)| The Pagerduty Provider for Selefra can be used to extract data from many of the cloud services by Pagerduty. The provider must be configured with credentials to extract and analyze infrastructure data from Pagerduty.    | Stable |
| [AliCloud](https://www.selefra.io/docs/providers-connector/alicloud)| The AliCloud Provider for Selefra can be used to extract data from many of the cloud services by AliCloud. The provider must be configured with credentials to extract and analyze infrastructure data from AliCloud.    | Stable |
| [Okta](https://www.selefra.io/docs/providers-connector/okta)| The Okta Provider for Selefra can be used to extract data from many of the cloud services by Okta. The provider must be configured with credentials to extract and analyze infrastructure data from Okta.    | Stable |
| [Snowflake](https://www.selefra.io/docs/providers-connector/snowflake)| The Snowflake Provider for Selefra can be used to extract data from many of the cloud services by Snowflake. The provider must be configured with credentials to extract and analyze infrastructure data from Snowflake.    | coming soon |

## Community

Selefra is a community-driven project, we welcome you to open a [GitHub Issue](https://github.com/selefra/selefra/issues/new/choose) to report a bug, suggest an improvement, or request new feature.

-  Join <a href="https://selefra.io/community/join"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225563969-3f3d4c45-fb3f-4932-831d-01ab9e59c921.png"></a> [Selefra Community](https://selefra.io/community/join) on Slack. We host `Community Hour` for tutorials and Q&As on regular basis.
-  Follow us on <a href="https://twitter.com/SelefraCorp"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225564426-82f5afbc-5638-4123-871d-fec6fdc6457f.png"></a> [Twitter](https://twitter.com/SelefraCorp) and share your thoughts！
-  Email us at <a href="support@selefra.io"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225564710-741dc841-572f-4cde-853c-5ebaaf4d3d3c.png"></a>&nbsp;support@selefra.io

## Contributing

For developers interested in building Selefra codebase, read through [Contributing.md](https://github.com/selefra/selefra/blob/main/CONTRIBUTING.md) and [Selefra Roadmap](https://github.com/orgs/selefra/projects/1).
Let us know what you would like to work on!

## License

[Mozilla Public License v2.0](https://github.com/selefra/selefra/blob/main/LICENSE)
