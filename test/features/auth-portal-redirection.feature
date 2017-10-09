Feature: User is redirected when factors are already validated
  
  @need-registered-user-john
  Scenario: User has validated first factor and tries to access service protected by second factor. He is then redirect to second factor step.
    When I visit "https://basicauth.test.local:8080/secret.html"
    And I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fbasicauth.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    And I'm redirected to "https://basicauth.test.local:8080/secret.html"
    And I visit "https://public.test.local:8080/secret.html"
    Then I'm redirected to "https://auth.test.local:8080/secondfactor?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"

  @need-registered-user-john
  Scenario: User who has validated second factor and access auth portal should be redirected to "Already logged in page"
    When I visit "https://public.test.local:8080/secret.html"
    And I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    And I'm redirected to "https://public.test.local:8080/secret.html"
    And I visit "https://auth.test.local:8080"
    Then I'm redirected to "https://auth.test.local:8080/loggedin"

  @need-registered-user-john
  Scenario: User who has validated second factor and access auth portal with rediction param should be redirected to that URL
    When I visit "https://public.test.local:8080/secret.html"
    And I'm redirected to "https://auth.test.local:8080/?redirect=https%3A%2F%2Fpublic.test.local%3A8080%2Fsecret.html"
    And I login with user "john" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    And I'm redirected to "https://public.test.local:8080/secret.html"
    And I visit "https://auth.test.local:8080?redirect=https://public.test.local:8080/secret.html"
    Then I'm redirected to "https://public.test.local:8080/secret.html"