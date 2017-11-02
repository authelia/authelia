Feature: User has access restricted access to domains

  @need-registered-user-john
  Scenario: User john has admin access
    When I visit "https://login.example.com:8080?redirect=https%3A%2F%2Fhome.example.com%3A8080%2F"
    And I login with user "john" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    And I'm redirected to "https://home.example.com:8080/"
    Then I have access to:
      | url                                                    |
      | https://public.example.com:8080/secret.html             |
      | https://dev.example.com:8080/groups/admin/secret.html   |
      | https://dev.example.com:8080/groups/dev/secret.html     |
      | https://dev.example.com:8080/users/john/secret.html     |
      | https://dev.example.com:8080/users/harry/secret.html    |
      | https://dev.example.com:8080/users/bob/secret.html      |
      | https://admin.example.com:8080/secret.html              |
      | https://mx1.mail.example.com:8080/secret.html           |
      | https://single_factor.example.com:8080/secret.html      |
    And I have no access to:
      | url                                                    |
      | https://mx2.mail.example.com:8080/secret.html           |

  @need-registered-user-bob
  Scenario: User bob has restricted access
    When I visit "https://login.example.com:8080?redirect=https%3A%2F%2Fhome.example.com%3A8080%2F"
    And I login with user "bob" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    And I'm redirected to "https://home.example.com:8080/"
    Then I have access to:
      | url                                                    |
      | https://public.example.com:8080/secret.html             |
      | https://dev.example.com:8080/groups/dev/secret.html     |
      | https://dev.example.com:8080/users/bob/secret.html      |
      | https://mx1.mail.example.com:8080/secret.html           |
      | https://mx2.mail.example.com:8080/secret.html           |
    And I have no access to:
      | url                                                    |
      | https://dev.example.com:8080/groups/admin/secret.html   |
      | https://admin.example.com:8080/secret.html              |
      | https://dev.example.com:8080/users/john/secret.html     |
      | https://dev.example.com:8080/users/harry/secret.html    |
      | https://single_factor.example.com:8080/secret.html      |

  @need-registered-user-harry
  Scenario: User harry has restricted access
    When I visit "https://login.example.com:8080?redirect=https%3A%2F%2Fhome.example.com%3A8080%2F"
    And I login with user "harry" and password "password"
    And I use "REGISTERED" as TOTP token handle
    And I click on "Sign in"
    And I'm redirected to "https://home.example.com:8080/"
    Then I have access to:
      | url                                                    |
      | https://public.example.com:8080/secret.html             |
      | https://dev.example.com:8080/users/harry/secret.html    |
    And I have no access to:
      | url                                                    |
      | https://dev.example.com:8080/groups/dev/secret.html     |
      | https://dev.example.com:8080/users/bob/secret.html      |
      | https://dev.example.com:8080/groups/admin/secret.html   |
      | https://admin.example.com:8080/secret.html              |
      | https://dev.example.com:8080/users/john/secret.html     |
      | https://mx1.mail.example.com:8080/secret.html           |
      | https://mx2.mail.example.com:8080/secret.html           |
      | https://single_factor.example.com:8080/secret.html      |