# One-Time Passwords

In **Authelia**, your users can use [Google Authenticator] for generating unique
tokens that they can use to pass the second factor.

<p align="center">
  <img src="../../docs/images/2FA-TOTP.png" width="400">
</p>

Select the *One-Time Password method* and click on the *Not registered yet?* link.
Then, check the email sent by **Authelia** to your email address
to validate your identity. If you're testing **Authelia**, it's likely
that this e-mail has been sent to https://mail.example.com:8080/

Confirm your identity by clicking on **Register** and you'll get redirected
on a page where your secret will be displayed as QRCode that you can scan.

<p align="center">
  <img src="../../docs/images/REGISTER-TOTP.png" width="400">
</p>

You can use [Google Authenticator] to store it.

From now on, you'll get tokens generated every 30 seconds on your phone that
you can use to validate the second factor in **Authelia**.



[Google Authenticator]: https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en