# Dockerize go service along with Nginx and PostgreSQL

Docker will make your life a bit easier when it comes to deployment and CI/CD. This method can be used to deploy most stacks with Nginx and Postgres,

## Installation

Your system must have [docker-compose](https://docs.docker.com/compose/install/) to follow along.

```bash
docker-compose build
docker-compose up
```
You would be able to access

[localhost:8008](http://localhost:8008/)

## Usage
stop the container
```bash
docker-compose down
```
drop to go service host shell
```bash
docker-compose exec web /bin/bash
```
more at [here](https://docs.docker.com/get-started/overview/)

## go service

Go service is solving following problem:
____________________________
As part of the assignment, we are sending you a $50 virtual card. Please reply stating that you agree to continue in the interview process and we will send you a virtual card for which you will receive an email notification. Note that you will be required to create an account on our website or mobile application in order to access it. This virtual card will be the source of the data you will be querying from our API for this project. To generate data, please make at least two purchases on this card; whatever you buy is yours to keep -- a token of appreciation for your efforts.


Using the virtual card we sent you, and the credentials associated with your user, we would like you to build an API service that exposes the following functionality:


1 - List all virtual cards available to your user, including the available balance remaining. In this case, we would expect to see just one virtual card returned.

2 - List the transactions associated with your virtual card.

3 - View the details for each individual transaction you’ve made.


Our API responses contain many fields. For each of the endpoints you expose, please return a “lite” view picking just a few important fields that demonstrate the main pieces of each response.


While this homework can be done in 2-3 hours, you are welcome to spend more time on it in order to demonstrate your technical ability.  In the follow-up interview, we will discuss your solution with you.  You will be expected to explain your thought process on a number of topics, including but not limited to: incomplete work, future enhancements, and implementation trade-offs. Some of the things we look for in the assignment are code clarity, code structure, and extensibility. Choices such as programming language, framework, or response formats are yours to make.


This is the type of work we do often at Extend: reading documents sent from card networks (such as, Visa, Mastercard, American Express), understanding the best way to perform similar operations on virtual cards, and returning that information to our frontend clients.  When working with these APIs, we often have questions about the documentation and/or functionality. If you come across something that you do not understand, please do not hesitate to reach out to myself or the engineering team at Extend for any clarification.

___________________________________
