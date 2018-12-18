Feature: User is able to reset his password

  Scenario: User is redirected to password reset page
    Given I'm on "https://login.example.com:8080"
    When I click on the link "Forgot password?"
    Then I'm redirected to "https://login.example.com:8080/password-reset/request"

  Scenario: User get an email with a link to reset password
    Given I'm on "https://login.example.com:8080/password-reset/request"
    When I set field "username" to "james"
    And I click on "Reset Password"
    Then I get a notification of type "success" with message "An email has been sent to you. Follow the link to change your password."

  Scenario: Request password for unexisting user should behave like existing user
    Given I'm on "https://login.example.com:8080/password-reset/request"
    When I set field "username" to "fake_user"
    And I click on "Reset Password"
    Then I get a notification of type "success" with message "An email has been sent to you. Follow the link to change your password."

  Scenario: User resets his password
    Given I'm on "https://login.example.com:8080/password-reset/request"
    And I set field "username" to "james"
    And I click on "Reset Password"
    When I click on the link of the email
    And I set field "password1" to "newpassword"
    And I set field "password2" to "newpassword"
    And I click on "Reset Password"
    Then I'm redirected to "https://login.example.com:8080/"


  Scenario: User does not confirm new password
    Given I'm on "https://login.example.com:8080/password-reset/request"
    And I set field "username" to "james"
    And I click on "Reset Password"
    When I click on the link of the email
    And I set field "password1" to "newpassword"
    And I set field "password2" to "newpassword2"
    And I click on "Reset Password"
    Then I get a notification of type "warning" with message "The passwords are different."
