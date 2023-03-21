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
<img alt="go" src="https://img.shields.io/badge/go-1.19-1E90FF"></a>
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
  <img src="https://user-images.githubusercontent.com/124020340/225897757-188f1a50-2efa-4a9e-9199-7cb7f68485be.png">
</p>
<br/>

## Contributing to Selefra

Welcome aboard Selefra! First thing first, thank you for contributing to selefra! 

### Code of Conduct 

We value each and every member, make sure to take some time and read the [Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md) to help maintain a productive and friendly community.

### Selefra Architecture

#### Overview

Selefra is a project that consists of four main components: CLI, SDK, Provider, and Module. The CLI provides the interface for users to interact with the Selefra system. The SDK provides the necessary tools and capabilities for the CLI and Provider to communicate with each other. The Provider is responsible for detecting certain conditions or issues, while the Module stores the standards and rules for the Provider to use.

- CLI

  The CLI component is the user-facing part of the Selefra system. It provides a command-line interface for users to interact with the system and run various commands. The CLI communicates with the SDK to initiate detection processes and receive results from the Provider.

- SDK

  The SDK is the foundation of the Selefra project, providing the necessary tools and capabilities for the CLI and Provider to communicate with each other. It includes a set of APIs and libraries that allow developers to integrate their applications with Selefra easily. The SDK also handles all the communication and data transfer between the CLI and the Provider.

- Provider

  The Provider is responsible for detecting specific conditions or issues and reporting them back to the CLI. It interacts with the SDK and the Module to perform its tasks. The Provider is designed to be modular and extensible, meaning that developers can add their own detection capabilities by creating new Provider modules.

- Module

  The Module is where the detection standards and rules are stored. It provides the data that the Provider needs to perform its tasks. The Module is designed to be flexible, allowing developers to customize or extend it based on their needs.

### Provider Registry

The Provider Registry is a Registry Services directory of domain-specific services that are exposed by products. If you want us to add a new plugin or resource please open an [Issue](https://github.com/selefra/selefra/issues).

### Submitting PR

1. Fork the repository to your own GitHub account.
2. Clone the project to your machine.
3. Create a new branch to work on. Branch from `develop` if it exists, else from `main`.
4. Make your changes and commit them. Make sure your commits are concise and descriptive.
5. Push your changes to your GitHub account.
6. Create a Pull Request (PR) from your branch to the `develop` branch in the main repository.


## Community

Selefra is a community-driven project, we welcome you to open a [GitHub Issue](https://github.com/selefra/selefra/issues/new/choose) to report a bug, suggest an improvement, or request new feature.

-  Join <a href="https://selefra.io/community/join"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225563969-3f3d4c45-fb3f-4932-831d-01ab9e59c921.png"></a> [Selefra Community](https://selefra.io/community/join) on Slack. We host `Community Hour` for tutorials and Q&As on regular basis.
-  Follow us on <a href="https://twitter.com/SelefraCorp"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225564426-82f5afbc-5638-4123-871d-fec6fdc6457f.png"></a> [Twitter](https://twitter.com/SelefraCorp) and share your thoughts！
-  Email us at <a href="support@selefra.io"><img height="16" alt="humanitarian" src="https://user-images.githubusercontent.com/124020340/225564710-741dc841-572f-4cde-853c-5ebaaf4d3d3c.png"></a>&nbsp;support@selefra.io

## License

[Mozilla Public License v2.0](https://github.com/selefra/selefra/blob/main/LICENSE)

