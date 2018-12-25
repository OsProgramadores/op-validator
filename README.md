# op-validator

## Synopsis

A challenge/exercise validator for the programming challenges in
[osprogramadores.com](https://osprogramadores.com).

## Usage

This program has three main handlers:

* The root handler (/), which serves a simple web page where where users can
  choose the desired challenge and type their github usernames and their
  program's output. On submit, the page makes an XMLHttpRequest to the
  check handler to verify the validity of the user's input.

* The check handler (/check) take the challenge, username and user's solution
  and attempts to match the solution to the canonical solution set in the
  config file.  If the user's input is valid, the program generates a token
  from the input data + a secret and instructs the user to add this token to a
  file called `.valid` in their repository before submitting a pull request.

* The verify token handler (/verify-token) takes the challenge, username, and a
  token, and verifies the validity of this token. The program returns an HTTP
  200 if the token is OK or HTTP 400 otherwise. The purpose of this handler is
  to allow remote token verification.
