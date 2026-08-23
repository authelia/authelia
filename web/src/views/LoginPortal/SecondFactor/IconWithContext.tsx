import { ReactNode } from "react";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    return (
        <div className={props.className}>
            <div className="flex flex-col items-center">
                <div className="flex size-17.5 items-center justify-center">{props.icon}</div>
            </div>
            <div className="block">{props.children}</div>
        </div>
    );
};

export default IconWithContext;
