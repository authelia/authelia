import { type ComponentProps } from "react";

import { Check } from "lucide-react";

import { cn } from "@utils/Styles";

interface StepperProps extends ComponentProps<"div"> {
    activeStep: number;
    children: React.ReactNode;
}

function Stepper({ activeStep, children, className, ...props }: StepperProps) {
    const steps = Array.isArray(children) ? children : [children];

    return (
        <div data-slot="stepper" className={cn("flex w-full items-center", className)} {...props}>
            {steps.map((child, index) => (
                <div key={index} className="flex flex-1 items-center">
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
                        {child}
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

interface StepProps extends ComponentProps<"div"> {
    completed?: boolean;
}

function Step({ children, className, ...props }: StepProps) {
    return (
        <div data-slot="step" className={cn(className)} {...props}>
            {children}
        </div>
    );
}

function StepLabel({ children, className, ...props }: ComponentProps<"span">) {
    return (
        <span data-slot="step-label" className={cn("mt-1 text-xs text-muted-foreground", className)} {...props}>
            {children}
        </span>
    );
}

export { Step, StepLabel, Stepper };
