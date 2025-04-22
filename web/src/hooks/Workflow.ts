import { useSearchParams } from "react-router-dom";

export function useWorkflow(): { id?: string; workflow?: string } {
    const [searchParams] = useSearchParams();

    const workflow = searchParams.get("workflow");
    const id = searchParams.get("workflow_id");

    return { id: id === null ? undefined : id, workflow: workflow === null ? undefined : workflow };
}
