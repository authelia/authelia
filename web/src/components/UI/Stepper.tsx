import { type ComponentProps } from "react";

import { Check } from "lucide-react";

import { cn } from "@utils/Styles";

interface StepperProps extends Omit<ComponentProps<"div">, "children"> {
    activeStep: number;
    steps: readonly string[];
}

function Stepper({ activeStep, className, steps, ...props }: Readonly<StepperProps>) {
    return (
        <div data-slot="stepper" className={cn("flex w-full items-center justify-center", className)} {...props}>
            {steps.map((label, index) => (
                <div key={label} className={cn("flex items-center", index < steps.length - 1 && "flex-1")}>
                    <div className="flex flex-col items-center gap-1">
                        <div
                            className={cn(
                                "flex size-8 items-center justify-center rounded-full border-2 text-sm font-medium transition-colors",
                                index < activeStep && "border-primary bg-primary text-primary-foreground",
                                index === activeStep && "border-primary bg-background text-primary",
                                index > activeStep &&
                                    "border-muted-foreground/40 bg-background text-muted-foreground/60",
                            )}
                        >
                            {index < activeStep ? <Check className="size-4" /> : index + 1}
                        </div>
                        <span data-slot="step-label" className="mt-1 text-xs text-muted-foreground">
                            {label}
                        </span>
                    </div>
                    {index < steps.length - 1 && (
                        <div
                            className={cn(
                                "mx-2 h-0.5 flex-1",
                                index < activeStep ? "bg-primary" : "bg-muted-foreground/30",
                            )}
                        />
                    )}
                </div>
            ))}
        </div>
    );
}

export { Stepper };
