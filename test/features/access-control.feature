Feature: User has access restricted access to domains

  @need-registered-user-john
  Scenario: User john has admin access
    When I visit "https://auth.test.local:8080"
    And I login with user "john" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                                    |
      | https://public.test.local:8080/secret.html             |
      | https://dev.test.local:8080/groups/admin/secret.html   |
      | https://dev.test.local:8080/groups/dev/secret.html     |
      | https://dev.test.local:8080/users/john/secret.html     |
      | https://dev.test.local:8080/users/harry/secret.html    |
      | https://dev.test.local:8080/users/bob/secret.html      |
      | https://admin.test.local:8080/secret.html              |
      | https://mx1.mail.test.local:8080/secret.html           |
      | https://basicauth.test.local:8080/secret.html          |
    And I have no access to:
      | url                                                    |
      | https://mx2.mail.test.local:8080/secret.html           |

  @need-registered-user-bob
  Scenario: User bob has restricted access
    When I visit "https://auth.test.local:8080"
    And I login with user "bob" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                                    |
      | https://public.test.local:8080/secret.html             |
      | https://dev.test.local:8080/groups/dev/secret.html     |
      | https://dev.test.local:8080/users/bob/secret.html      |
      | https://mx1.mail.test.local:8080/secret.html           |
      | https://mx2.mail.test.local:8080/secret.html           |
    And I have no access to:
      | url                                                    |
      | https://dev.test.local:8080/groups/admin/secret.html   |
      | https://admin.test.local:8080/secret.html              |
      | https://dev.test.local:8080/users/john/secret.html     |
      | https://dev.test.local:8080/users/harry/secret.html    |
      | https://basicauth.test.local:8080/secret.html          |

  @need-registered-user-harry
  Scenario: User harry has restricted access
    When I visit "https://auth.test.local:8080"
    And I login with user "harry" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "TOTP"
    Then I have access to:
      | url                                                    |
      | https://public.test.local:8080/secret.html             |
      | https://dev.test.local:8080/users/harry/secret.html    |
    And I have no access to:
      | url                                                    |
      | https://dev.test.local:8080/groups/dev/secret.html     |
      | https://dev.test.local:8080/users/bob/secret.html      |
      | https://dev.test.local:8080/groups/admin/secret.html   |
      | https://admin.test.local:8080/secret.html              |
      | https://dev.test.local:8080/users/john/secret.html     |
      | https://mx1.mail.test.local:8080/secret.html           |
      | https://mx2.mail.test.local:8080/secret.html           |
      | https://basicauth.test.local:8080/secret.html          |
