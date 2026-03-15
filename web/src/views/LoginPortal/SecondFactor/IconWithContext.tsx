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
                <div className="flex h-16 w-16 items-center justify-center">{props.icon}</div>
            </div>
            <div className="block">{props.children}</div>
        </div>
    );
};

export default IconWithContext;
