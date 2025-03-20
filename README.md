# Pulumi ESC: Simplifying Secrets and Configurations
Pulumi ESC (Environments, Secrets, and Configurations) provides a developer-first solution to simplify how you manage sensitive data and configurations across your entire application lifecycle. It's a fully managed solution allowing teams to generate dynamic cloud provider credentials, aggregate secrets, and configurations from multiple sources, and manage them through composable collections called "environments." These environments can be consumed from anywhere, making Pulumi ESC ideal for any application and development workflow. Additionally, while Pulumi ESC works independently to eliminate duplication and reduce drift and sprawl of secrets and configuration for all your applications, it also integrates smoothly with Pulumi Infrastructure as Code (IaC) to enhance these capabilities within the Pulumi ecosystem.


### Introducing the Pulumi ESC SDK
We're excited to unveil the official Pulumi ESC SDK, making it even easier to harness the power of ESC directly within your applications using your favorite programming languages. The SDK provides a simple, programmatic interface to all of Pulumi ESC's robust features, allowing you to:

Manage the Entire Lifecycle of Your Environments: Create new environments, list existing ones, and easily update or delete them as your needs evolve. You can even add version tags to your environments, making it simple to track changes and roll back to previous states if needed.
Seamlessly Integrate Secrets and Configurations: Securely access and utilize secrets and configurations within your applications. The SDK provides a streamlined way to fetch the information you need, whether it's cloud credentials, database connection strings, feature flags, or any other sensitive data.


### Why use the Pulumi ESC SDK?
Here's why you'll love using the Pulumi ESC SDK:

* Focus on Building, Not Plumbing: No more writing boilerplate code for API interactions. The SDK handles the complexities of secret retrieval and configuration management, so you can focus on your application's core functionality.
* Eliminate Hardcoded Secrets: Say goodbye to hardcoded credentials scattered throughout your codebase. With the SDK, your application can securely retrieve secrets from Pulumi ESC at runtime.
* Centralized Configuration Management: Manage all your application settings from a single source of truth. The SDK provides easy access to configurations stored in Pulumi ESC, ensuring consistency across environments.
* Enhanced Security: The Pulumi ESC SDK promotes secure secret handling best practices. It handles the secure storage and retrieval of your sensitive data, minimizing the risk of accidental exposure.
* Intuitive and Idiomatic: The SDK is built to feel like a natural extension of your chosen programming language. You'll work with familiar objects, methods, and patterns, making integration smooth and intuitive.
* Type Safety and Code Completion: Benefit from the power of your IDE. The SDK provides type safety, enabling compile-time checks and helpful code suggestions that reduce errors and speed up your development workflow.

## Getting Started

### TypeScript/JavaScript

```shell
npm install @pulumi/esc-sdk
```

[TypeScript/JavaScript examples](https://www.pulumi.com/docs/esc/development/languages-sdks/javascript/)

### Python

```shell
pip install pulumi-esc-sdk
```

[Python examples](https://www.pulumi.com/docs/esc/development/languages-sdks/python/)

### Go

```shell
go get github.com/pulumi/esc-sdk/sdk
```

[Go examples](https://www.pulumi.com/docs/esc/development/languages-sdks/go/)

## API Reference Documentation

* [TypeScript/JavaScript](https://www.pulumi.com/docs/reference/pkg/nodejs/pulumi/esc-sdk/)
* [Python](https://www.pulumi.com/docs/reference/pkg/python/pulumi_esc_sdk/)
* [Go](https://pkg.go.dev/github.com/pulumi/esc-sdk/sdk/go)
