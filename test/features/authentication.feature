Feature: User validate first factor

  Scenario: User succeeds first factor
    Given I visit "https://auth.test.local:8080/"
    When I set field "username" to "bob"
    And I set field "password" to "password"
    And I click on "Sign in"
    Then I'm redirected to "https://auth.test.local:8080/secondfactor"

  Scenario: User fails first factor
    Given I visit "https://auth.test.local:8080/"
    When I set field "username" to "john"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    Then I get a notification with message "Error during authentication: Authetication failed. Please check your credentials."

  Scenario: User succeeds TOTP second factor
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://secret.test.local:8080/secret.html" and get redirected "https://auth.test.local:8080/"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://secret.test.local:8080/secret.html"

  Scenario: User fails TOTP second factor
    When I visit "https://secret.test.local:8080/secret.html" and get redirected "https://auth.test.local:8080/"
    And I login with user "john" and password "password" 
    And I use "BADTOKEN" as TOTP token
    And I click on "TOTP"
    Then I get a notification with message "Error while validating TOTP token. Cause: error"

  Scenario: User logs out
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    And I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    When I visit "https://auth.test.local:8080/logout?redirect=https://www.google.fr"
    And I visit "https://secret.test.local:8080/secret.html"
    Then I'm redirected to "https://auth.test.local:8080/"

  Scenario: Logout redirects user
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    And I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    When I visit "https://auth.test.local:8080/logout?redirect=https://www.google.fr"
    Then I'm redirected to "https://www.google.fr"