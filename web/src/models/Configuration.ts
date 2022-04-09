import { SecondFactorMethod } from "@models/Methods";

export interface Configuration {
    available_methods: Set<SecondFactorMethod>;
}

export interface LocalesConfiguration {
    supported: string[];
}
