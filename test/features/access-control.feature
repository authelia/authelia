Feature: User has access restricted access to domains

  @need-registered-user-john
  Scenario: User john has admin access
    When I visit "https://auth.test.local:8080"
    And I login with user "john" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |
      | https://secret.test.local:8080/secret.html   |
      | https://secret1.test.local:8080/secret.html  |
      | https://secret2.test.local:8080/secret.html  |
      | https://mx1.mail.test.local:8080/secret.html |
      | https://mx2.mail.test.local:8080/secret.html |

  @need-registered-user-bob
  Scenario: User bob has restricted access
    When I visit "https://auth.test.local:8080"
    And I login with user "bob" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |
      | https://secret.test.local:8080/secret.html   |
      | https://secret2.test.local:8080/secret.html  |
      | https://mx1.mail.test.local:8080/secret.html |
      | https://mx2.mail.test.local:8080/secret.html |
    And I have no access to:
      | url                                          |
      | https://secret1.test.local:8080/secret.html  |

  @need-registered-user-harry
  Scenario: User harry has restricted access
    When I visit "https://auth.test.local:8080"
    And I login with user "harry" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |
      | https://secret1.test.local:8080/secret.html  |
    And I have no access to:
      | url                                          |
      | https://secret.test.local:8080/secret.html   |
      | https://secret2.test.local:8080/secret.html  |
      | https://mx1.mail.test.local:8080/secret.html |
      | https://mx2.mail.test.local:8080/secret.html |