# Introduction

Core Blockchain aspires to engage with responsible global security experts to bolster the security of the Core Blockchain ecosystem. We've instituted a program to facilitate the reporting of vulnerabilities to the Core Team and to recognize your efforts in strengthening blockchain as a reliable and trustworthy technology.

## Security Declaration

The Core development team's allocated time frame for implementing a solution varies according to the severity level indicated in the report, which can last up to 90 days. Kindly ensure this process has been concluded before bringing the vulnerability to public attention.

## Rewards

In recognition of researchers uncovering systemic flaws, we extend bounties in our digital currency, Core Coin (XCB). The final amount of the reward is contingent upon the severity level of the reported vulnerability.

The assessment of rewards in our system aligns with the [OWASP](https://www.owasp.org/index.php/OWASP_Risk_Rating_Methodology) risk rating model, which takes into account both impact and likelihood in the calculation process.

| **Impact**   | **Severity**    | **Severity**    | **Severity**    |
|--------------|-----------------|-----------------|-----------------|
| **High**     | Moderate //S3// | High //S4//     | Critical //S5// |
| **Moderate** | Low //S2//      | Moderate //S3// | High //S4//     |
| **Low**      | Note //S1//     | Low //S2//      | Moderate //S3// |
|              | **Low**         | **Medium**      | **High**        |
|              | **Likelihood**  | **Likelihood**  | **Likelihood**  |

## Eligibility

In our assessment, a bug's potential to jeopardize the security or stability of the Core Network forms the basis for potential reward qualification. However, it ultimately falls under our purview to discern whether a bug aligns with the established criteria for reward eligibility.

- You must be the very first individual to ethically report a previously unknown problem.
- The bug must be reported using the most recent software version.
- Ensure the bug has not been previously identified. Please, review any [published security advisories](https://github.com/core-coin/go-core/security/advisories).

Generally, the following vulnerabilities are not considered severe and are, therefore, not eligible for reporting:

- 0-day vulnerabilities that have just been released.
- Vulnerabilities reliant on physical attacks, spamming, DDoS attacks, social engineering, etc.
- Vulnerabilities on third-party hosted sites that are not proven prone to causing a vulnerability on a main website scale.
- Third-party application vulnerabilities utilizing Core Blockchain’s API.
- Vulnerabilities found on past versions of or otherwise unpatched applications.
- Flaws that have not been thoroughly investigated and have not been reported in a satisfactory manner.
- Issues with no successful reproducibility.
- Bugs that the team had prior knowledge of or those previously reported by another party (the reward is granted to the initial reporter).
- Vulnerabilities that the team cannot reasonably be expected to address.

## Severity

The evaluation of a bug's severity plays a pivotal role in determining the monetary reward. Factors taken into account encompass the number of individuals affected and whether the core or supplementary modules are implicated.

## Best Practices

- Employ your localized go-core instance alongside a distinct network (avoid the test or public networks) for discovering security vulnerabilities.
- Keep in mind the public nature of blockchains; someone might come across your discoveries and report a bug ahead of you.
- Providing a detailed step-by-step report or an exploit script is highly encouraged. This expedites our comprehension and resolution of the issue, ensuring you receive your rewards promptly.

## Legal

Residents or individuals situated within countries listed on any EU sanctions roster are ineligible to participate in this program. It is incumbent upon you to account for any tax ramifications or supplementary constraints contingent on your country's specific laws and regulations. While we reserve the right to amend or conclude this program at any time, any alterations made to the program terms will not be applied retroactively.

## Vulnerability Reporting

- Ensure your report is comprehensive, encompassing details about the bug's nature, potential consequences, and instructions for reproduction or a proof of concept.
- Express your message in English.
- Kindly allow a period of three business days for our response before considering sending a follow-up email.

[Report Vulnerability](https://dev.coreblockchain.net/vulnerability-report)
