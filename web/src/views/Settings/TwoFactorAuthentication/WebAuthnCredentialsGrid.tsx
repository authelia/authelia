import { WebAuthnCredential } from "@models/WebAuthn";
import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

interface Props {
    credentials: WebAuthnCredential[];
    handleInformation: (_index: number) => void;
    handleEdit: (_index: number) => void;
    handleDelete: (_index: number) => void;
}

const WebAuthnCredentialsGrid = function (props: Props) {
    return (
        <div className="grid grid-cols-12 gap-6">
            {props.credentials.map((credential, index) => (
                <div className="col-span-12 md:col-span-6 xl:col-span-3" key={credential.id}>
                    <WebAuthnCredentialItem
                        index={index}
                        credential={credential}
                        handleInformation={props.handleInformation}
                        handleEdit={props.handleEdit}
                        handleDelete={props.handleDelete}
                    />
                </div>
            ))}
        </div>
    );
};

export default WebAuthnCredentialsGrid;
