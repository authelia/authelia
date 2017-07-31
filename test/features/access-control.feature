Feature: User has access restricted access to domains

  Scenario: User john has admin access
    When I register TOTP and login with user "john" and password "password"
    Then I have access to:
      | url                                          |
      | https://public.test.local:8080/secret.html   |
      | https://secret.test.local:8080/secret.html   |
      | https://secret1.test.local:8080/secret.html  |
      | https://secret2.test.local:8080/secret.html  |
      | https://mx1.mail.test.local:8080/secret.html |
      | https://mx2.mail.test.local:8080/secret.html |

  Scenario: User bob has restricted access
    When I register TOTP and login with user "bob" and password "password"
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

  Scenario: User harry has restricted access
    When I register TOTP and login with user "harry" and password "password"
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