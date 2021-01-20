package templates

import (
	"text/template"
)

// HTMLEmailTemplate the template of email that the user will receive for identity verification.
var HTMLEmailTemplate *template.Template

func init() {
	t, err := template.New("html_email_template").Parse(emailHTMLContent)
	if err != nil {
		panic(err)
	}

	HTMLEmailTemplate = t
}

const emailHTMLContent = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
   <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
   <meta name="viewport" content="width=device-width, initial-scale=1.0" />
   <title>Authelia</title>

   <style type="text/css">
      /* client-specific Styles */
      #outlook a {
         padding: 0;
      }

      /* Force Outlook to provide a "view in browser" menu link. */
      body {
         width: 100% !important;
         -webkit-text-size-adjust: 100%;
         -ms-text-size-adjust: 100%;
         margin: 0;
         padding: 0;
      }

      /* Prevent Webkit and Windows Mobile platforms from changing default font sizes, while not breaking desktop design. */
      .ExternalClass {
         width: 100%;
      }

      /* Force Hotmail to display emails at full width */
      .ExternalClass,
      .ExternalClass p,
      .ExternalClass span,
      .ExternalClass font,
      .ExternalClass td,
      .ExternalClass div {
         line-height: 100%;
      }

      /* Force Hotmail to display normal line spacing.*/
      #backgroundTable {
         margin: 0;
         padding: 0;
         width: 100% !important;
         line-height: 100% !important;
      }

      img {
         outline: none;
         text-decoration: none;
         border: none;
         -ms-interpolation-mode: bicubic;
      }

      a img {
         border: none;
      }

      .image_fix {
         display: block;
      }

      p {
         margin: 0px 0px !important;
      }

      table td {
         border-collapse: collapse;
      }

      table {
         border-collapse: collapse;
         mso-table-lspace: 0pt;
         mso-table-rspace: 0pt;
      }

      a {
         color: #0a8cce;
         text-decoration: none;
         text-decoration: none !important;
      }

      h1 {
         line-height: 30px;
      }

      .button {
         padding: 15px 30px;
         border-radius: 10px;
         background: rgb(204, 204, 255);
         text-decoration: none;
      }

      /*STYLES*/
      table[class=full] {
         width: 100%;
         clear: both;
      }

      /*IPAD STYLES*/
      @media only screen and (max-width: 640px) {

         a[href^="tel"],
         a[href^="sms"] {
            text-decoration: none;
            color: #0a8cce;
            /* or whatever your want */
            pointer-events: none;
            cursor: default;
         }

         .mobile_link a[href^="tel"],
         .mobile_link a[href^="sms"] {
            text-decoration: default;
            color: #0a8cce !important;
            pointer-events: auto;
            cursor: default;
         }

         table[class=devicewidth] {
            width: 440px !important;
            text-align: center !important;
         }

         table[class=devicewidthinner] {
            width: 420px !important;
            text-align: center !important;
         }

         img[class=banner] {
            width: 440px !important;
            height: 220px !important;
         }

         img[class=colimg2] {
            width: 440px !important;
            height: 220px !important;
         }

      }

      /*IPHONE STYLES*/
      @media only screen and (max-width: 480px) {

         a[href^="tel"],
         a[href^="sms"] {
            text-decoration: none;
            color: #0a8cce;
            /* or whatever your want */
            pointer-events: none;
            cursor: default;
         }

         .mobile_link a[href^="tel"],
         .mobile_link a[href^="sms"] {
            text-decoration: default;
            color: #0a8cce !important;
            pointer-events: auto;
            cursor: default;
         }

         table[class=devicewidth] {
            width: 280px !important;
            text-align: center !important;
         }

         table[class=devicewidthinner] {
            width: 260px !important;
            text-align: center !important;
         }

         img[class=banner] {
            width: 280px !important;
            height: 140px !important;
         }

         img[class=colimg2] {
            width: 280px !important;
            height: 140px !important;
         }

         td[class=mobile-hide] {
            display: none !important;
         }

         td[class="padding-bottom25"] {
            padding-bottom: 25px !important;
         }

      }
   </style>
</head>

<body>
   <!-- Start of header -->
   <table width="100%" bgcolor="#ffffff" cellpadding="0" cellspacing="0" border="0" id="backgroundTable"
      st-sortable="header">
      <tbody>
         <tr>
            <td>
               <table width="600" cellpadding="0" cellspacing="0" border="0" align="center" class="devicewidth">
                  <tbody>
                     <tr>
                        <td width="100%">
                           <table width="600" cellpadding="0" cellspacing="0" border="0" align="center"
                              class="devicewidth">
                              <tbody>
                                 <!-- Spacing -->
                                 <tr>
                                    <td height="20"
                                       style="font-size:1px; line-height:1px; mso-line-height-rule: exactly;">&nbsp;
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                                 <tr>
                                    <td>
                                       <!-- logo -->
                                       <table width="140" align="center" border="0" cellpadding="0" cellspacing="0"
                                          class="devicewidth">
                                          <tbody>
                                             <tr>
                                                <td width="300" height="50" align="center">
                                                   <h1>{{.title}}</h1>
                                                </td>
                                             </tr>
                                          </tbody>
                                       </table>
                                       <!-- end of logo -->
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                                 <tr>
                                    <td height="20"
                                       style="font-size:1px; line-height:1px; mso-line-height-rule: exactly;">&nbsp;
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                              </tbody>
                           </table>
                        </td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
   <!-- End of Header -->
   <!-- Start of separator -->
   <table width="100%" bgcolor="#ffffff" cellpadding="0" cellspacing="0" border="0" id="backgroundTable"
      st-sortable="separator">
      <tbody>
         <tr>
            <td>
               <table width="600" align="center" cellspacing="0" cellpadding="0" border="0" class="devicewidth">
                  <tbody>
                     <tr>
                        <td align="center" height="20" style="font-size:1px; line-height:1px;">&nbsp;</td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
   <!-- End of separator -->
   <!-- Start Full Text -->
   <table width="100%" bgcolor="#ffffff" cellpadding="0" cellspacing="0" border="0" id="backgroundTable"
      st-sortable="full-text">
      <tbody>
         <tr>
            <td>
               <table width="600" cellpadding="0" cellspacing="0" border="0" align="center" class="devicewidth">
                  <tbody>
                     <tr>
                        <td width="100%">
                           <table width="600" cellpadding="0" cellspacing="0" border="0" align="center"
                              class="devicewidth">
                              <tbody>
                                 <!-- Spacing -->
                                 <tr>
                                    <td height="20"
                                       style="font-size:1px; line-height:1px; mso-line-height-rule: exactly;">&nbsp;
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                                 <tr>
                                    <td>
                                       <table width="560" align="center" cellpadding="0" cellspacing="0" border="0"
                                          class="devicewidthinner">
                                          <tbody>
                                             <!-- Title -->
                                             <tr>
                                                <td style="font-family: Helvetica, arial, sans-serif; font-size: 16px; color: #333333; text-align:center; line-height: 30px;"
                                                   st-title="fulltext-content">
                                                   This email has been sent to you in order to validate your identity.
                                                   If you did not initiate the process your credentials might have been compromised. You should reset your password and contact an administrator.
                                                </td>
                                             </tr>
                                             <!-- End of Title -->
                                             <!-- spacing -->
                                             <tr>
                                                <td width="100%" height="20"
                                                   style="font-size:1px; line-height:1px; mso-line-height-rule: exactly;">
                                                   &nbsp;</td>
                                             </tr>
                                             <!-- End of spacing -->
                                             <!-- content -->
                                             <tr>
                                                <td style="font-family: Helvetica, arial, sans-serif; font-size: 16px; color: #666666; text-align:center; line-height: 30px;"
                                                   st-content="fulltext-content">
                                                   <a href="{{.url}}" class="button">{{.button}}</a>
                                                </td>
                                             </tr>
                                             <!-- End of content -->
                                          </tbody>
                                       </table>
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                                 <tr>
                                    <td height="20"
                                       style="font-size:1px; line-height:1px; mso-line-height-rule: exactly;">&nbsp;
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                              </tbody>
                           </table>
                        </td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
   <!-- end of full text -->
   <!-- Start of separator -->
   <table width="100%" bgcolor="#ffffff" cellpadding="0" cellspacing="0" border="0" id="backgroundTable"
      st-sortable="separator">
      <tbody>
         <tr>
            <td>
               <table width="600" align="center" cellspacing="0" cellpadding="0" border="0" class="devicewidth">
                  <tbody>
                     <tr>
                        <td align="center" height="30" style="font-size:1px; line-height:1px;">&nbsp;</td>
                     </tr>
                     <tr>
                        <td width="550" align="center" height="1" bgcolor="#d1d1d1"
                           style="font-size:1px; line-height:1px;">&nbsp;</td>
                     </tr>
                     <tr>
                        <td align="center" height="30" style="font-size:1px; line-height:1px;">&nbsp;</td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
   <!-- End of separator -->
   <!-- Start of Postfooter -->
   <table width="100%" bgcolor="#ffffff" cellpadding="0" cellspacing="0" border="0" id="backgroundTable"
      st-sortable="postfooter">
      <tbody>
         <tr>
            <td>
               <table width="600" cellpadding="0" cellspacing="0" border="0" align="center" class="devicewidth">
                  <tbody>
                     <tr>
                        <td width="100%">
                           <table width="600" cellpadding="0" cellspacing="0" border="0" align="center"
                              class="devicewidth">
                              <tbody>
                                 <tr>
                                    <td align="center" valign="middle"
                                       style="font-family: Helvetica, arial, sans-serif; font-size: 14px;color: #666666"
                                       st-content="postfooter">
                                       Please ignore this email if you did not initiate the process.
                                    </td>
                                 </tr>
                                 <!-- Spacing -->
                                 <tr>
                                    <td width="100%" height="20"></td>
                                 </tr>
                                 <!-- Spacing -->
                              </tbody>
                           </table>
                        </td>
                     </tr>
                  </tbody>
               </table>
            </td>
         </tr>
      </tbody>
   </table>
   <!-- End of postfooter -->
</body>

</html>
`
