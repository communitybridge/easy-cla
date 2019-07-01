# Easy Contributor License Agreement (CLA) Project

[![CircleCI](https://circleci.com/gh/communitybridge/easy-cla.svg?style=svg&circle-token=936fbcd340fff1b26453a2f46ffeab48f4fd3be4)](https://circleci.com/gh/communitybridge/easy-cla)

The Contributor License Agreement (CLA) service of the Linux Foundation lets
project contributors read, sign, and submit contributor license agreements easily.

 This repository contains both the backend and front-end UI for supporting and
 managing the application.

This platform supports both GitHub and Gerrit source code repositories.
Additional information can be found in the [User's Guide](#users-guide).

## Users Guide

A user's guide can be found
[on the Contributor License Agreement public confluence pages](https://docs.linuxfoundation.org/display/DOCS/Contributor+License+Agreement+User+Guide).

## Third-party Services

Besides integration with Auth0 and Salesforce, the CLA system has the following third party services:

- [Docusign](https://www.docusign.com/) for CLA agreement e-sign flow
- [Docraptor](https://docraptor.com/) for convert html CLA template as PDF file

## CLA Backend

The CLA project has two backend projects.

The majority of the backend APIs are implemented in python, and can be found
in the `cla-backend` directory.

Recent backend development was implemented in Golang, and can be found in the
`cla-backend-go` directory. In particular, this backend contains APIs
powering Automated Templates, GitHub Whitelists, and Duplicate Company
handling in the Corporate Console.

## CLA Frontend

### Overview

CLA frontend consists of three independent SPA build with
[Ionic](https://ionicframework.com/) framework.

- `frontend-project-management-console` for LinuxFoundation director/admin/user to manage project CLA
- `cla-frontend-corporate-console` for any concrete company CLA manager to sign CCLA and manage employee CLA whitelist
- `cla-frontend-console` for any project contributor to sign ICLA or CCLA

## License

Copyright The Linux Foundation and each contributor to CommunityBridge.

This project’s source code is licensed under the MIT License. A copy of the
license is available in LICENSE-code.txt.

It includes source code from keycloak, which is licensed under the Apache License,
version 2.0 (Apache-2.0), a copy of which is also available in LICENSE-code.txt.

This project’s documentation is licensed under the Creative Commons Attribution
4.0 International License (CC-BY-4.0). A copy of the license is available in
LICENSE-docs.txt.

