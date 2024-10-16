# Phones Tester

This implements a server that will manage mobile phones issued by AWS User Messaging Service. It is designed for use as a test tool to receive messages sent to an AWS User Messaging phone and store those messages in memory then provide the user with a REST API for viewing and managing those messages,  including sending replies.

This can be deployed in an EKS cluster and used to receive the messages sent by the HOSPENG messaging service to AWS phones to for testing.

## Design Notes

Refer to [Design Guide](docs/design-guide.md)

## Developers Guide

Refer to [Developer's Guide](docs/dev-guide.md)

## Version History

### Version 0.0.1

Initial version
