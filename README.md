# op-validator

## Synopsis

A challenge/exercise validator for the programming challenges in
[osprogramadores.com](https://osprogramadores.com).

## Usage

This program has two main operational modes: Web/API server and standalone.

As a web server, it serves a simple page where users can choose the desired
challenge and type their github usernames and their program's output. It will
then proceed to check the validity of the user's output against a canonical
value set in a configuration file. If the user's input is valid, the program
generates a token from the input data + a secret and instructs the user to add
this token to a file called `.valid` in their repository before submitting a
pull request.

As a standalone program, it checks validity of the token in the `.valid` file.
This allows github continuous integration frameworks like
[CircleCI](http://circleci.com) or [TravisCI](http://travisci.com) to block the
merge of invalid solutions.
