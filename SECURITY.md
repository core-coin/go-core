## Introduction

We in Core Blockchain want to engage with responsible security researchers around the globe to further secure the Blockchain environment.
We developed a program to make it easier to report vulnerabilities to the Core Team and to recognize your efforts to make the Blockchain a secure and reliable technology.

## Responsible Disclosure

The Core development team has up to 90 days to implement a fix based on the severity of the report. Please allow for this process to be fully completed before you publicly disclose the vulnerability.

## Rewards

We are rewarding researchers that find bugs with a bounty of our digital currency, Core Coin (XCB). The amount of the award depends on the degree of severity of the vulnerability reported.

We calculate rewards accordingly to the [OWASP](https://www.owasp.org/index.php/OWASP_Risk_Rating_Methodology) risk rating model based on Impact and Likelihood.

| **Impact**   | **Severity**    | **Severity**    | **Severity**    |
|--------------|-----------------|-----------------|-----------------|
| **High**     | Moderate //S3// | High //S4//     | Critical //S5// |
| **Moderate** | Low //S2//      | Moderate //S3// | High //S4//     |
| **Low**      | Note //S1//     | Low //S2//      | Moderate //S3// |
|              | **Low**         | **Medium**      | **High**        |
|              | **Likelihood**  | **Likelihood**  | **Likelihood**  |

## Eligibility

Any bug that poses a significant vulnerability to the security or integrity of the Core Network could be eligible for a reward. However, it’s entirely at our discretion to decide whether a bug is significant enough to be eligible for a reward.

- You must be the first person to responsibly disclose an unknown issue.
- You must report the bug via the latest version of the software.
- Bug shouldn't be already discovered. Please, check [published security advisories](https://github.com/core-coin/go-core/security/advisories).

In general, the following would not meet the threshold for severity:

- Recently disclosed 0-day vulnerabilities
- Vulnerabilities on sites hosted by third parties, unless they lead to a vulnerability on the main website
- Vulnerabilities contingent on physical attack, social engineering, spamming, DDOS attack, etc.
- Vulnerabilities affecting outdated or unpatched browsers
- Vulnerabilities in third-party applications that make use of Stellar’s API
- Bugs that have not been responsibly investigated and reported
- Bugs already known to us, or already reported by someone else (reward goes to the first reporter)
- Issues that aren’t reproducible
- Issues that we can’t reasonably be expected to do anything about

## Severity

The severity of a bug is taken into consideration when deciding the bounty payout amount. We consider how many people are affected as well if the core or additional modules are affected.

## Best Practices

- Please use your local instance of Go-core and a separate network (not test/public network) when searching for security bugs. Remember that blockchains are public and someone may see your findings and report a bug before you.
- Step by step report (or an exploit script) is more than welcomed. It will allow us to understand and fix the issue faster and you will get your rewards more quickly.

## Legal

You may not participate in this program if you are a resident or individual located within a country appearing on any EU sanctions lists.
You are responsible for any tax implications or additional restrictions depending on your country and local law.
We may modify the terms of this program or terminate this program at any time, but we won’t apply any changes we make to these program terms retroactively.

## Report Vulnerability

* Try to include as much information in your report as you can, including a description of the bug, its potential impact, and steps for reproducing it or proof of concept.
* Compose your message in the English language.
* Please allow 3 business days for us to respond before sending another email.

[Report Vulnerability](https://dev.coreblockchain.cc/vulnerability-report)
