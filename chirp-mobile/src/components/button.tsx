import { Button as HeroButton, type ButtonRootProps } from "heroui-native/button";

/*
 * App-owned Button: a string label renders single-line (trailing ellipsis)
 * instead of wrapping, so an action's text always fits the button plate.
 */

export const Button = Object.assign(
  function Button(props: ButtonRootProps) {
    const { children, ...rest } = props;
    const isTextLabel = typeof children === "string" || typeof children === "number";

    return (
      <HeroButton {...rest}>
        {isTextLabel ? <HeroButton.Label numberOfLines={1}>{children}</HeroButton.Label> : children}
      </HeroButton>
    );
  },
  { Label: HeroButton.Label },
);
