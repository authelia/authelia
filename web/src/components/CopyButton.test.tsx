import React from "react";

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, vi } from "vitest";

import CopyButton from "@components/CopyButton";

const mockWriteText = vi.fn(() => Promise.resolve());

vi.stubGlobal("navigator", {
    clipboard: {
        writeText: mockWriteText,
    },
});

beforeEach(() => {
    mockWriteText.mockClear();
});

it("renders without crashing", () => {
    render(
        <CopyButton tooltip="copy" value="test">
            Copy
        </CopyButton>,
    );
});

it("renders disabled button when value is null", () => {
    render(
        <CopyButton tooltip="copy" value={null}>
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    expect(button).toBeDisabled();
});

it("renders disabled button when value is empty", () => {
    render(
        <CopyButton tooltip="copy" value="">
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    expect(button).toBeDisabled();
});

it("copies to clipboard on click", async () => {
    render(
        <CopyButton tooltip="copy" value="test">
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockWriteText).toHaveBeenCalledWith("test");
});

it("shows copied text after copying", async () => {
    render(
        <CopyButton tooltip="copy" value="test" childrenCopied="Copied">
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText("Copy")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Copied")).toBeInTheDocument(), { timeout: 600 });
    await waitFor(() => expect(screen.getByText("Copy")).toBeInTheDocument(), { timeout: 2100 });
});

it("does not copy if value is null", () => {
    render(
        <CopyButton tooltip="copy" value={null}>
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockWriteText).not.toHaveBeenCalled();
});

it("does not copy if value is empty", () => {
    render(
        <CopyButton tooltip="copy" value="">
            Copy
        </CopyButton>,
    );
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockWriteText).not.toHaveBeenCalled();
});
