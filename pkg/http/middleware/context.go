package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/nano-interactive/GoFiber-Boilerplate/pkg/constants"
)

func Context(ctx *fiber.Ctx) error {
	c, cancel := context.WithCancel(context.Background())

	ctx.Locals(constants.CancelFuncContextKey, cancel)
	ctx.SetUserContext(c)

	err := ctx.Next()

	cancelFnWillBeCalled := ctx.Locals(constants.CancelWillBeCalledContextKey)

	if cancelFnWillBeCalled == nil {
		defer cancel()
	}

	return err
}
