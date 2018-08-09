Feature: User is correctly redirected 

  Scenario: User is redirected to authelia when he is not authenticated
    When I visit "https://public.example.com:8080"
    Then I'm redirected to "https://login.example.com:8080/?rd=https://public.example.com:8080/"

  @need-registered-user-john
  Scenario: User is redirected to home page after several authentication tries
    When I visit "https://public.example.com:8080/secret.html"
    And I login with user "john" and password "badpassword"
    And I wait for notification to disappear
    And I clear field "username"
    And I clear field "password"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://public.example.com:8080/secret.html"

  Scenario: User Harry does not have access to admin domain and thus he must get an error 403
    When I register TOTP and login with user "harry" and password "password"
    And I visit "https://admin.example.com:8080/secret.html"
    Then I get an error 403

  Scenario: Redirection URL is propagated from restricted page to first factor
    When I visit "https://public.example.com:8080/secret.html"
    Then I'm redirected to "https://login.example.com:8080/?rd=https://public.example.com:8080/secret.html"

  Scenario: Redirection URL is propagated from first factor to second factor
    Given I visit "https://login.example.com:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://public.example.com:8080/secret.html"
    And I login with user "john" and password "password"
    Then I'm redirected to "https://login.example.com:8080/secondfactor?rd=https://public.example.com:8080/secret.html"

  Scenario: Redirection URL is used to send user from second factor to target page
    Given I visit "https://login.example.com:8080/"
    And I login with user "john" and password "password"
    And I register a TOTP secret called "Sec0"
    When I visit "https://public.example.com:8080/secret.html"
    And I login with user "john" and password "password"
    And I use "Sec0" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://public.example.com:8080/secret.html"

  @need-registered-user-john
  Scenario: User is redirected to default URL defined in configuration when authentication is successful
    When I visit "https://login.example.com:8080"
    And I login with user "john" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    Then I'm redirected to "https://home.example.com:8080/"


  Scenario: User is redirected when hitting an error 401
    When I visit "https://login.example.com:8080/secondfactor/u2f/identity/finish"
    Then I'm redirected to "https://login.example.com:8080/error/401"
    And I sleep for 5 seconds
    And I'm redirected to "https://home.example.com:8080/"

  @need-registered-user-harry
  Scenario: User is redirected when hitting an error 403
    When I visit "https://login.example.com:8080"
    And I login with user "harry" and password "password" 
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    And I'm redirected to "https://home.example.com:8080/"
    When I visit "https://admin.example.com:8080/secret.html"
    Then I'm redirected to "https://login.example.com:8080/error/403"
    And I sleep for 5 seconds
    And I'm redirected to "https://home.example.com:8080/"
