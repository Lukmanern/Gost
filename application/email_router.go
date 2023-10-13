// don't use this for production
// used for sending testing email in development

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
	emailRoute := router.Group("email")
	emailRoute.Post("send", func(c *fiber.Ctx) error {
		err := emailService.Send("lukmanernandi16@gmail.com", "Testing Gost Project", simpleMessage)
		if err != nil {
			return base.ResponseInternalServerError(c, "internal server error: "+err.Error())
		}

		return base.ResponseUpdated(c, "success sending message")
	})
}
