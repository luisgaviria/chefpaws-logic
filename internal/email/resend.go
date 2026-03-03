package email

import (
	"fmt"
	"os"

	"github.com/luisgaviria/chefpaws-logic/internal/models"
	resend "github.com/resend/resend-go/v2"
)

// SendContactEmail emails the team when the contact form is submitted.
func SendContactEmail(apiKey, name, emailAddr, breed, message string) error {
	client := resend.NewClient(apiKey)

	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "onboarding@resend.dev"
	}
	to := os.Getenv("NOTIFICATION_EMAIL")
	if to == "" {
		to = "luis.aptx@gmail.com"
	}

	breedLine := ""
	if breed != "" {
		breedLine = fmt.Sprintf(`
        <tr>
          <td style="padding:0 36px 20px;">
            <div style="background:#f7f8f9;border-radius:10px;padding:14px 16px;">
              <p style="margin:0 0 4px 0;font-size:10px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#9e9e9e;">Breed</p>
              <p style="margin:0;font-size:16px;font-weight:700;color:#1a1a1a;">%s</p>
            </div>
          </td>
        </tr>`, breed)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background-color:#f2f4f6;font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f2f4f6;padding:40px 16px;">
    <tr><td align="center">
      <table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">

        <!-- Header -->
        <tr>
          <td style="background-color:#111111;padding:32px 36px;">
            <p style="margin:0 0 4px 0;font-size:11px;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:rgba(255,255,255,0.5);">Contact Form — ChefPaws</p>
            <h1 style="margin:0;font-size:26px;font-weight:800;color:#ffffff;line-height:1.2;">%s</h1>
            <p style="margin:6px 0 0 0;font-size:14px;color:rgba(255,255,255,0.7);">%s</p>
          </td>
        </tr>

        %s

        <!-- Divider -->
        <tr><td style="padding:0 36px;"><div style="height:1px;background:#eeeeee;"></div></td></tr>

        <!-- Message -->
        <tr>
          <td style="padding:24px 36px 28px;">
            <p style="margin:0 0 10px 0;font-size:10px;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:#9e9e9e;">Message</p>
            <div style="background:#fffbf0;border-left:3px solid #ff6b6b;border-radius:0 8px 8px 0;padding:14px 16px;">
              <p style="margin:0;font-size:14px;color:#333333;line-height:1.7;">%s</p>
            </div>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="background:#f7f8f9;padding:16px 36px;border-top:1px solid #eeeeee;">
            <p style="margin:0;font-size:11px;color:#bdbdbd;">Source: ChefPaws Contact Form</p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>
</body>
</html>`, name, emailAddr, breedLine, message)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: fmt.Sprintf("📬 Contact from %s — ChefPaws", name),
		Html:    html,
	}

	resp, err := client.Emails.Send(params)
	fmt.Printf("📬 Resend Contact Response: %+v\n", resp)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}
	return nil
}

// SendLeadNotification emails the team when a new lead is captured.
// Logs the full Resend API response so errors are visible in the terminal.
func SendLeadNotification(apiKey string, lead models.Lead) error {
	client := resend.NewClient(apiKey)

	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "ChefPaws <onboarding@resend.dev>"
	}

	to := os.Getenv("NOTIFICATION_EMAIL")
	if to == "" {
		to = "luis.aptx@gmail.com"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background-color:#f2f4f6;font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;">

  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f2f4f6;padding:40px 16px;">
    <tr><td align="center">

      <!-- Card -->
      <table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">

        <!-- Header -->
        <tr>
          <td style="background-color:#43A047;padding:32px 36px;">
            <p style="margin:0 0 4px 0;font-size:11px;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:rgba(255,255,255,0.75);">New Lead — ChefPaws Calculator</p>
            <h1 style="margin:0;font-size:28px;font-weight:800;color:#ffffff;line-height:1.2;">%s</h1>
            <p style="margin:6px 0 0 0;font-size:15px;color:rgba(255,255,255,0.85);">%s &nbsp;·&nbsp; %s</p>
          </td>
        </tr>

        <!-- Stats grid -->
        <tr>
          <td style="padding:28px 36px 20px;">
            <p style="margin:0 0 16px 0;font-size:10px;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:#9e9e9e;">Nutrition Profile</p>
            <table width="100%%" cellpadding="0" cellspacing="0">
              <tr>
                <td width="50%%" style="padding-bottom:16px;padding-right:12px;">
                  <div style="background:#f7f8f9;border-radius:10px;padding:14px 16px;">
                    <p style="margin:0 0 4px 0;font-size:10px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#9e9e9e;">Daily Calories</p>
                    <p style="margin:0;font-size:22px;font-weight:800;color:#1a1a1a;">%.0f <span style="font-size:12px;font-weight:600;color:#9e9e9e;">kcal</span></p>
                  </div>
                </td>
                <td width="50%%" style="padding-bottom:16px;">
                  <div style="background:#f7f8f9;border-radius:10px;padding:14px 16px;">
                    <p style="margin:0 0 4px 0;font-size:10px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#9e9e9e;">Portion Size</p>
                    <p style="margin:0;font-size:22px;font-weight:800;color:#1a1a1a;">%.0f <span style="font-size:12px;font-weight:600;color:#9e9e9e;">g/day</span></p>
                  </div>
                </td>
              </tr>
              <tr>
                <td width="50%%" style="padding-right:12px;">
                  <div style="background:#f7f8f9;border-radius:10px;padding:14px 16px;">
                    <p style="margin:0 0 4px 0;font-size:10px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#9e9e9e;">Phone</p>
                    <p style="margin:0;font-size:16px;font-weight:700;color:#1a1a1a;">%s</p>
                  </div>
                </td>
                <td width="50%%">
                  <div style="background:#f7f8f9;border-radius:10px;padding:14px 16px;">
                    <p style="margin:0 0 4px 0;font-size:10px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#9e9e9e;">Zip Code</p>
                    <p style="margin:0;font-size:16px;font-weight:700;color:#1a1a1a;">%s</p>
                  </div>
                </td>
              </tr>
            </table>
          </td>
        </tr>

        <!-- Divider -->
        <tr><td style="padding:0 36px;"><div style="height:1px;background:#eeeeee;"></div></td></tr>

        <!-- Special Requirements -->
        <tr>
          <td style="padding:20px 36px 28px;">
            <p style="margin:0 0 10px 0;font-size:10px;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:#9e9e9e;">Special Requirements</p>
            <div style="background:#fffbf0;border-left:3px solid #43A047;border-radius:0 8px 8px 0;padding:14px 16px;">
              <p style="margin:0;font-size:14px;color:#333333;line-height:1.6;">%s</p>
            </div>
          </td>
        </tr>

        <!-- Footer -->
        <tr>
          <td style="background:#f7f8f9;padding:16px 36px;border-top:1px solid #eeeeee;">
            <p style="margin:0;font-size:11px;color:#bdbdbd;">Source: %s</p>
          </td>
        </tr>

      </table>
    </td></tr>
  </table>

</body>
</html>
	`,
		lead.DogName, lead.OwnerName, lead.Email,
		lead.DailyCalories, lead.PortionGrams,
		lead.Phone, lead.Zip,
		lead.SpecialReqs,
		lead.Source,
	)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: fmt.Sprintf("🐾 New Lead: %s's owner — %s", lead.DogName, lead.OwnerName),
		Html:    html,
	}

	resp, err := client.Emails.Send(params)
	fmt.Printf("📬 Resend Response: %+v\n", resp)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}
	return nil
}
