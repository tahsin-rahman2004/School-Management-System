"use strict";
(() => {
  const subjects = [
    { name: "Bangla", marks: 85, grade: "A+", points: 5 },
    { name: "English", marks: 78, grade: "A", points: 4 },
    { name: "Mathematics", marks: 92, grade: "A+", points: 5 },
    { name: "Science", marks: 88, grade: "A+", points: 5 },
    { name: "BGS", marks: 80, grade: "A+", points: 5 },
    { name: "Religion", marks: 85, grade: "A+", points: 5 }
  ];
  const total = subjects.reduce((sum, subject) => sum + subject.marks, 0);
  const gpa = (subjects.reduce((sum, subject) => sum + subject.points, 0) / subjects.length).toFixed(2);
  const form = document.getElementById("resultSearch");
  const sheet = document.getElementById("marksheet");
  const actions = document.getElementById("resultActions");
  const empty = document.getElementById("resultEmpty");
  const message = document.getElementById("searchMessage");
  let resultVisible = false;
  document.getElementById("subjectRows").replaceChildren(...subjects.map(subject => {
    const row = document.createElement("tr");
    [subject.name, subject.marks, subject.grade].forEach(value => {
      const cell = document.createElement("td");
      cell.textContent = value;
      row.append(cell);
    });
    return row;
  }));
  document.getElementById("totalMarks").textContent = `${total}/600`;
  document.getElementById("resultGpa").textContent = gpa;
  function hideResult() {
    resultVisible = false;
    sheet.hidden = true;
    actions.hidden = true;
    empty.hidden = false;
    document.getElementById("downloadStatus").textContent = "";
  }
  function showResult() {
    resultVisible = true;
    sheet.hidden = false;
    actions.hidden = false;
    empty.hidden = true;
    message.textContent = "Sample result loaded. This is not a real student record.";
    sheet.focus({ preventScroll: true });
    sheet.scrollIntoView({ behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth", block: "start" });
  }
  form.addEventListener("submit", event => {
    event.preventDefault();
    const validDemo = document.getElementById("resultYear").value === "2026" &&
      document.getElementById("resultClass").value === "8" &&
      document.getElementById("resultExam").value === "annual" &&
      document.getElementById("resultRoll").value.trim() === "12";
    hideResult();
    if (validDemo) showResult();
    else message.textContent = "No demo result for these details. Try 2026, Class 8, Annual Examination, roll 12. Real records are not connected.";
  });
  form.addEventListener("input", () => { hideResult(); message.textContent = ""; });
  form.addEventListener("change", () => { hideResult(); message.textContent = ""; });
  document.getElementById("loadDemo").addEventListener("click", () => {
    document.getElementById("resultYear").value = "2026";
    document.getElementById("resultClass").value = "8";
    document.getElementById("resultExam").value = "annual";
    document.getElementById("resultRoll").value = "12";
    showResult();
  });
  document.getElementById("printResult").addEventListener("click", () => {
    if (resultVisible) window.print();
  });
  // Dependency-free single-page PDF. All content is ASCII, so string offsets
  // equal PDF byte offsets. The sample label is included in the exported file.
  function createSamplePdf() {
    const escapeText = value => String(value).replace(/([\\()])/g, "\\$1");
    const commands = [];
    const text = (x, y, size, value) => commands.push(`BT /F1 ${size} Tf ${x} ${y} Td (${escapeText(value)}) Tj ET`);
    commands.push("0.09 0.18 0.29 rg");
    text(48, 785, 20, "S. M. Pilot Model High School");
    text(48, 756, 12, "Annual Examination - 2026 | ACADEMIC TRANSCRIPT");
    commands.push("0.65 0.15 0.22 rg");
    text(48, 725, 12, "SAMPLE RESULT - NOT AN OFFICIAL RECORD");
    commands.push("0.09 0.18 0.29 rg");
    text(48, 684, 11, "Student Name: Demo Student");
    text(335, 684, 11, "Class / Roll: 8 / 12");
    text(48, 662, 11, "Section: A");
    text(335, 662, 11, "Academic Year: 2026");
    text(48, 620, 11, "Subject"); text(345, 620, 11, "Marks / 100"); text(465, 620, 11, "Grade");
    commands.push("0.75 0.8 0.86 RG 48 608 m 547 608 l S");
    subjects.forEach((subject, index) => {
      const y = 583 - index * 32;
      text(48, y, 11, subject.name); text(370, y, 11, subject.marks); text(475, y, 11, subject.grade);
    });
    commands.push("48 398 m 547 398 l S");
    text(48, 367, 14, `Total: ${total}/600`); text(252, 367, 14, `GPA: ${gpa}`); text(420, 367, 14, "Passed");
    text(48, 315, 10, "Illustrative marks and GPA only; actual grades must follow");
    text(48, 298, 10, "the school's approved grading rules.");
    text(48, 250, 10, "Demo reference: SAMPLE-2026-8-12");
    text(48, 230, 10, "This preview is not connected to a real student database.");
    text(48, 70, 10, "SAMPLE ONLY - NOT VALID FOR OFFICIAL USE");
    const stream = commands.join("\n") + "\n";
    const objects = [
      "<< /Type /Catalog /Pages 2 0 R >>",
      "<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
      "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
      "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
      `<< /Length ${stream.length} >>\nstream\n${stream}endstream`
    ];
    let pdf = "%PDF-1.4\n";
    const offsets = [0];
    objects.forEach((object, index) => { offsets.push(pdf.length); pdf += `${index + 1} 0 obj\n${object}\nendobj\n`; });
    const xref = pdf.length;
    pdf += `xref\n0 ${objects.length + 1}\n0000000000 65535 f \n`;
    offsets.slice(1).forEach(offset => { pdf += `${String(offset).padStart(10, "0")} 00000 n \n`; });
    pdf += `trailer\n<< /Size ${objects.length + 1} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`;
    return new Blob([pdf], { type: "application/pdf" });
  }
  document.getElementById("downloadResult").addEventListener("click", () => {
    if (!resultVisible) return;
    const url = URL.createObjectURL(createSamplePdf());
    const link = document.createElement("a");
    link.href = url;
    link.download = "sample-marksheet-2026-class8-roll12.pdf";
    document.body.append(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 30000);
    document.getElementById("downloadStatus").textContent = "Sample PDF download requested.";
  });
  document.getElementById("verifyForm").addEventListener("submit", event => {
    event.preventDefault();
    const reference = document.getElementById("verifyReference").value.trim().toUpperCase();
    document.getElementById("verifyMessage").textContent = reference === "SAMPLE-2026-8-12"
      ? "Recognized demo reference only. This sample is not an authentic school-issued marksheet."
      : "Official verification is unavailable in this preview. Contact the school office to verify a real marksheet.";
  });
})();
