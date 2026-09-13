interface CsvColumn<T> {
    header: string;
    value: (row: T) => string;
}

function escapeCsvValue(value: string): string {
    if (/[",\n\r]/.test(value)) {
        return `"${value.replace(/"/g, '""')}"`;
    }

    return value;
}

function toCsv<T>(columns: CsvColumn<T>[], rows: T[]): string {
    const lines = [columns.map((column) => escapeCsvValue(column.header)).join(",")];

    for (const row of rows) {
        lines.push(columns.map((column) => escapeCsvValue(column.value(row))).join(","));
    }

    return lines.join("\r\n");
}

function downloadCsv(name: string, text: string): void {
    const blob = new Blob([text], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);

    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = name;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);

    URL.revokeObjectURL(url);
}

export { toCsv, downloadCsv, escapeCsvValue };
export type { CsvColumn };
