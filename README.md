# Firefly III BoC Fixer

## Problem

The Bank of Cyprus provides an open banking API through third-party vendors for integration with Firefly III. However, the quality of the integration is poor, and the data mapping is incorrect.

Below is a list of issues related to both the open banking API and the CSV export:

- **Date**: The provided dates are not the actual payment dates; they are likely future clearance dates.
- **Description**: The description is disorganized and difficult to read. It contains data that is randomly concatenated into a single string, such as: `Card 1***2345 2024-10-15 160.56 EUR Auth 123456 Trace 123456 PURCHASE CY CYTA BILLS`.

## Solution


```
Card 1***2345 2024-10-15 160.56 EUR Auth 123456 Trace 123456 PURCHASE CY CYTA BILLS
```

In the example above, the date in the description refers to when the payment was initiated. The order of data elements in the string may vary, and sometimes there is no space between "Auth" and the following number. However, it is possible to extract the transaction date and remove extraneous data using regular expressions. This project does exactly that, utilizing the Firefly API and webhooks to update transactions accordingly. From my personal experience, it handles 99% of BoC cases correctly.

## Setup

The application can be configured using environment variables:

| Variable                 | Description                               |
| ------------------------ | ----------------------------------------- |
| `FIREFLY_URL`            | Required. The Firefly host.               |
| `FIREFLY_TOKEN`          | Firefly API token                         |
| `WEBHOOK_SECRET`         | Firefly webhook secret                    |
| `FIREFLY_BOC_FIXER_HOST` | Host to listen on. Defaults to `0.0.0.0`. |
| `FIREFLY_BOC_FIXER_PORT` | Port to listen on. Defaults to `3000`.    |

Additionally, there is a `nix` package and module provided for convenience.
