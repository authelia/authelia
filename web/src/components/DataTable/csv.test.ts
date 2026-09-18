import { downloadCsv, escapeCsvValue, toCsv } from "@components/DataTable/csv";

interface Row {
    name: string;
    note: string;
}

const columns = [
    { header: "Name", value: (row: Row) => row.name },
    { header: "Note", value: (row: Row) => row.note },
];

it("escapes a value containing a comma", () => {
    expect(escapeCsvValue("a,b")).toBe('"a,b"');
});

it("escapes a value containing a quote by doubling it", () => {
    expect(escapeCsvValue('say "hi"')).toBe('"say ""hi"""');
});

it("escapes a value containing a newline", () => {
    expect(escapeCsvValue("line1\nline2")).toBe('"line1\nline2"');
});

it("does not escape a plain value", () => {
    expect(escapeCsvValue("plain")).toBe("plain");
});

it("builds a header row plus one row per record", () => {
    const rows: Row[] = [
        { name: "Alice", note: "ok" },
        { name: "Bob", note: "fine" },
    ];

    expect(toCsv(columns, rows)).toBe("Name,Note\r\nAlice,ok\r\nBob,fine");
});

it("escapes commas, quotes and newlines within generated rows", () => {
    const rows: Row[] = [{ name: "Doe, John", note: 'said "hello"\nagain' }];

    expect(toCsv(columns, rows)).toBe('Name,Note\r\n"Doe, John","said ""hello""\nagain"');
});

it("produces just the header row for an empty row list", () => {
    expect(toCsv(columns, [])).toBe("Name,Note");
});

it("downloadCsv creates an object URL, clicks a temporary anchor, then revokes the URL", () => {
    const createObjectURL = vi.fn(() => "blob:mock-url");
    const revokeObjectURL = vi.fn();
    URL.createObjectURL = createObjectURL;
    URL.revokeObjectURL = revokeObjectURL;

    const clickSpy = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => {});
    const appendSpy = vi.spyOn(document.body, "appendChild");
    const removeSpy = vi.spyOn(document.body, "removeChild");

    downloadCsv("export.csv", "a,b\r\n1,2");

    expect(createObjectURL).toHaveBeenCalledTimes(1);
    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(appendSpy).toHaveBeenCalled();
    expect(removeSpy).toHaveBeenCalled();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");

    clickSpy.mockRestore();
    appendSpy.mockRestore();
    removeSpy.mockRestore();
});
