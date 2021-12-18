---
layout: default
title: Password Policy
parent: Features
nav_order: 8
---

# Password Policy
Password policy enforces the security by requering the users to use strong passwords
Currently, two methods are supported:
## classic
* this mode of operation allows administrators to set the rules that user passwords must comply with
* the available options are: 
    * Minimum password length
    * Require Uppercase
    * Require Lowercase
    * Require Numbers
    * Require Special characters
* the password entered by the user must meet these rules 


<p align="center">
  <img src="../images/password-policy-classic-1.png" width="400">
</p>


## zxcvbn
* this mode uses zxcvbn for password strength checking (see: https://github.com/dropbox/zxcvbn)
* in this mode of operation, the user is not forced to follow any rules. the user is notified if their passwords is weak or strong
<p align="center">
  <img src="../images/password-policy-zxcvbn-1.png" width="400">
</p>




