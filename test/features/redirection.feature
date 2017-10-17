Feature: User is correctly redirected 

  Scenario: User is redirected to authelia when he is not authenticated
    When I visit "https://public.test.local:8080"
    Then I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2F"

  @need-registered-user-john
  Scenario: User is redirected to home page after several authentication tries
    When I visit "https://public.test.local:8080/secret.html"
    And I login with user "john" and password "badpassword"
    And I wait for notification to disappear
    And I clear field "username"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://public.test.local:8080/secret.html"

  Scenario: User Harry does not have access to admin domain and thus he must get an error 403
    When I register TOTP and login with user "harry" and password "password"
    And I visit "https://admin.test.local:8080/secret.html"
    Then I get an error 403

  Scenario: Redirection URL is propagated from restricted page to first factor
    When I visit "https://public.test.local:8080/secret.html"
    Then I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"

  Scenario: Redirection URL is propagated from first factor to second factor
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://public.test.local:8080/secret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://auth.test.local:8080/secondfactor?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"

  Scenario: Redirection URL is used to send user from second factor to target page
    Given I visit "https://auth.test.local:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://public.test.local:8080/secret.html"
    And I login with user "john" and password "password"
    And I use "Sec0" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://public.test.local:8080/secret.html"

  @need-registered-user-john
  Scenario: User is redirected to default URL defined in configuration when authentication is successful
    When I visit "https://auth.test.local:8080"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://home.test.local:8080/"