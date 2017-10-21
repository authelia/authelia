Feature: Authentication scenarii

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
    Then I get a notification of type "error" with message "Authentication failed. Please check your credentials."

  Scenario: User registers TOTP secret and succeeds authentication
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://admin.test.local:8080/secret.html"
    And I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://admin.test.local:8080/secret.html"

  Scenario: User fails TOTP second factor
    When I visit "https://admin.test.local:8080/secret.html"
    And I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fadmin.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    And I use "BADTOKEN" as TOTP token
    And I click on "Sign in"
    Then I get a notification of type "error" with message "Authentication failed. Have you already registered your secret?"

  Scenario: Logout redirects user to redirect URL given in parameter
    When I visit "https://auth.test.local:8080/logout?redirect=https://home.test.local:8080/"
    Then I'm redirected to "https://home.test.local:8080/"
