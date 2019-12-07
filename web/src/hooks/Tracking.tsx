import { useState, useEffect } from "react";
import { Configuration } from "../models/Configuration";
import { Tracker } from "react-ga";

export function useTracking(configuration: Configuration | undefined) {
    const [trackingIds, setTrackingIds] = useState(undefined as Tracker | undefined);

    useEffect(() => {
        if (configuration && configuration.ga_tracking_id) {
            setTrackingIds({ trackingId: configuration.ga_tracking_id });
        }
    }, [configuration]);

    return trackingIds;
}