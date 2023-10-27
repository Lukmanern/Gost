package controller

// func TestDelete(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/action/%s", actionEntity.ID), nil)
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
// 	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
// 	app := fiber.New()
// 	actionService := service.LoadActionService()
// 	actionController := NewActionController(actionService)
// 	app.Put("/action/:id", actionController.Delete)
// 	resp, err := app.Test(req, -1)
// 	if err != nil {
// 		log.Print(err)
// 		tearDown()
// 		return
// 	}
// 	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
// 	saved, _ := repo.Get(ctx, actionEntity.ID, hospitalID)
// 	assert.Empty(t, saved.ID)
// }
