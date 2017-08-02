Feature: User is correctly redirected correctly 

  Scenario: User is redirected to authelia when he is not authenticated
    Given I'm on https://home.test.local:8080
    When I click on the link to secret.test.local
    Then I'm redirected to "https://auth.test.local:8080/"

  Scenario: User is redirected to home page after several authentication tries
    Given I'm on https://auth.test.local:8080/
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    And I visit "https://public.test.local:8080/secret.html"
    When I login with user "john" and password "badpassword"
    And I clear field "username"
    And I login with user "john" and password "password" 
    And I use "Sec0" as TOTP token handle
    And I click on "TOTP"
    Then I'm redirected to "https://public.test.local:8080/secret.html"

  Scenario: User Harry does not have access to https://secret.test.local:8080/secret.html and thus he must get an error 401
    When I register TOTP and login with user "harry" and password "password"
    And I visit "https://secret.test.local:8080/secret.html"
    Then I get an error 403