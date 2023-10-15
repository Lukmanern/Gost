// don't use this for production
// used for sending email in development

// for dev : router -> email service (without controller)

package application

import (
	"github.com/Lukmanern/gost/domain/base"
	service "github.com/Lukmanern/gost/service/email"
	"github.com/gofiber/fiber/v2"
)

const simpleMessage = `Lorem ipsum, dolor sit amet consectetur adipisicing elit. Ad consequuntur 
similique voluptatibus ab enim harum dolor, sit, corporis repellendus culpa cum, quasi corrupti! 
Impedit inventore cum optio quas, nisi aliquid ullam omnis voluptas, architecto deserunt, sint 
tempora? Iure ea alias recusandae sunt ad, vero laudantium esse.`

var emailService service.EmailService

func getEmailRouter(router fiber.Router) {
	emailService = service.NewEmailService()
	emailRoutes := router.Group("email")
	emailRoutes.Post("send-bulk", func(c *fiber.Ctx) error {
		testEmails := []string{"lukmanernandi16@gmail.com", "unsurlukman@gmail.com", "code_name_safe_in_unsafe@proton.me", "lukmanernandi16@gmail.com.", "unsurlukm an@gmail.com", "code _name_safe_in_unsafe@proton.me", "lukmanern*a)ndi16@gmail.com", "unsurlukman@gmail.com", "code_n}ame_safe_in_unsafe@proton.me"}
		res, err := emailService.Send(testEmails, "Testing Gost Project", simpleMessage)
		if err != nil {
			return base.ResponseErrorWithData(c, "internal server error: "+err.Error(), fiber.Map{
				"res": res,
			})
		}
		if res == nil {
			return base.ResponseError(c, "internal server error: failed sending email")
		}

		message := "success sending emails"
		return base.Response(c, fiber.StatusAccepted, true, message, nil)
	})
}
