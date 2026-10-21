// Sample seat counts. Replace these with school-confirmed totals before opening admissions.
// 0 = no seats; null = not confirmed.
const CLASS_SEAT_AVAILABILITY = {
  2: 18, 3: 12, 4: 0, 5: 9, 6: 40,
  7: 15, 8: 0, 9: 8, 10: 5
};

function canStartApplication(value) {
  const rules = getAdmissionRequirements(value);
  return rules.valid && Number(value) !== 10 && (Number(value) === 1 || Number.isInteger(CLASS_SEAT_AVAILABILITY[Number(value)]) && CLASS_SEAT_AVAILABILITY[Number(value)] > 0);
}

function getAdmissionRequirements(value) {
  const classNumber = Number(value);
  const valid = Number.isInteger(classNumber) && classNumber >= 1 && classNumber <= 10;
  return { valid, needsAcademicCertificate: valid && classNumber >= 2 && classNumber <= 9, needsExam: valid && classNumber >= 2 && classNumber <= 9, selection: !valid ? null : classNumber === 1 ? 'lottery' : classNumber <= 9 ? 'written-exam' : 'school-notice' };
}

(() => {
  let language = 'en';
  const translate = (en, bn) => language === 'bn' ? bn : en;
  const nodes = [...document.querySelectorAll('[data-bn]')];
  nodes.forEach(node => { node.dataset.en = node.textContent; });
  document.getElementById('languageToggle').addEventListener('click', event => {
    language = language === 'en' ? 'bn' : 'en';
    document.documentElement.lang = language;
    nodes.forEach(node => { node.textContent = node.dataset[language]; });
    event.currentTarget.textContent = language === 'en' ? 'বাংলা' : 'English';
    renderSeatInformation();
    if (document.getElementById('applicationForm')) updateClassRequirements();
  });
  function renderSeatInformation() {
    const class10Seats = CLASS_SEAT_AVAILABILITY[10];
    const vacancy = document.getElementById('class10Vacancy');
    if (vacancy) vacancy.textContent = Number.isInteger(class10Seats) && class10Seats > 0
      ? translate(`${class10Seats} seats available in Class 10. Contact the school office for admission. There is no online application for Class 10.`, `১০ম শ্রেণিতে ${class10Seats}টি আসন খালি আছে। ভর্তির জন্য স্কুল অফিসে যোগাযোগ করুন। ১০ম শ্রেণির অনলাইন আবেদন নেই।`)
      : class10Seats === 0
        ? translate('No seats available in Class 10. Admission cannot proceed.', '১০ম শ্রেণিতে খালি আসন নেই। ভর্তি করা যাবে না।')
        : translate('Class 10 vacancy information is not confirmed. Contact the school office.', '১০ম শ্রেণির আসনসংখ্যা নিশ্চিত নয়। স্কুল অফিসে যোগাযোগ করুন।');
    document.querySelectorAll('[data-seat-class]').forEach(node => {
      const value = CLASS_SEAT_AVAILABILITY[Number(node.dataset.seatClass)];
      node.textContent = Number.isInteger(value) && value > 0 ? translate(`${value} seats`, `${value}টি আসন`) : value === 0 ? translate('No seats', 'খালি নেই') : translate('Not confirmed', 'নিশ্চিত নয়');
    });
  }
  renderSeatInformation();
  if (!document.getElementById('applicationForm')) return;
  const storageKey = 'school-admission-drafts-v1';
  function readDrafts() {
    const data = JSON.parse(localStorage.getItem(storageKey) || '{}');
    if (!data || typeof data !== 'object' || Array.isArray(data)) throw new Error('Invalid storage');
    return data;
  }
  const application = document.getElementById('applicationForm');
  const photoInput = document.getElementById('studentPhoto');
  const photoPreview = document.getElementById('studentPhotoPreview');
  const photoError = document.getElementById('photoError');
  let photoUrl = null;
  function clearPhotoPreview() {
    if (photoUrl) URL.revokeObjectURL(photoUrl);
    photoUrl = null;
    photoPreview.removeAttribute('src');
    photoPreview.hidden = true;
  }
  photoInput.addEventListener('change', () => {
    clearPhotoPreview();
    photoInput.setCustomValidity('');
    photoError.textContent = '';
    const file = photoInput.files[0];
    if (!file) return;
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type) || file.size > 5 * 1024 * 1024) {
      const error = translate('Choose a JPG, PNG or WebP photo no larger than 5 MB.', 'সর্বোচ্চ ৫ মেগাবাইটের JPG, PNG বা WebP ছবি দিন।');
      photoInput.setCustomValidity(error);
      photoError.textContent = error;
      photoInput.reportValidity();
      return;
    }
    photoUrl = URL.createObjectURL(file);
    photoPreview.src = photoUrl;
    photoPreview.hidden = false;
  });
  photoPreview.addEventListener('error', () => {
    clearPhotoPreview();
    const error = translate('This photo could not be opened. Choose another image.', 'ছবিটি খোলা যায়নি। অন্য ছবি দিন।');
    photoInput.setCustomValidity(error);
    photoError.textContent = error;
  });
  window.addEventListener('pagehide', event => { if (!event.persisted) clearPhotoPreview(); });
  const certificateInputs = ['transferCertificate', 'academicCertificate'].map(id => document.getElementById(id));
  function validateCertificate(input) {
    input.setCustomValidity('');
    const error = document.getElementById(input.id + 'Error');
    error.textContent = '';
    if (input.matches(':disabled')) return true;
    const file = input.files[0];
    if (!file) return false;
    if (!['application/pdf', 'image/jpeg', 'image/png', 'image/webp'].includes(file.type) || file.size > 5 * 1024 * 1024) {
      const text = translate('Choose a PDF, JPG, PNG or WebP file no larger than 5 MB.', 'সর্বোচ্চ ৫ মেগাবাইটের PDF, JPG, PNG বা WebP ফাইল দিন।');
      input.setCustomValidity(text);
      error.textContent = text;
      return false;
    }
    return true;
  }
  certificateInputs.forEach(input => input.addEventListener('change', () => validateCertificate(input)));


  document.getElementById('printApplication').addEventListener('click', () => {
    if (canStartApplication(application.elements.class.value)) window.print();
  });
  const message = document.getElementById('applicationMessage');
  const classSelect = application.elements.class;
  const academicFields = document.getElementById('academicFields');
  function updateClassRequirements() {
    const rules = getAdmissionRequirements(classSelect.value);
    const allowed = canStartApplication(classSelect.value);
    const details = document.getElementById('applicationDetails');
    details.hidden = !allowed;
    details.disabled = !allowed;
    if (document.activeElement === classSelect) message.hidden = true;
    const availability = CLASS_SEAT_AVAILABILITY[Number(classSelect.value)];
    const status = document.getElementById('seatAvailability');
    status.textContent = !rules.valid
      ? translate('Choose a class to continue.', 'এগিয়ে যেতে শ্রেণি নির্বাচন করুন।')
      : Number(classSelect.value) === 1
        ? translate('Class 1: continue to the lottery application form.', '১ম শ্রেণি: লটারির আবেদন পূরণ করুন।')
        : Number.isInteger(availability) && availability > 0
          ? Number(classSelect.value) === 10
            ? translate(`${availability} seats available. Contact the school office for Class 10 admission; no online application is offered.`, `${availability}টি আসন খালি আছে। ১০ম শ্রেণিতে ভর্তির জন্য স্কুল অফিসে যোগাযোগ করুন; অনলাইন আবেদন নেই।`)
            : translate(`${availability} seats available — you may continue.`, `${availability}টি আসন খালি আছে — আবেদন পূরণ করতে পারেন।`)
          : availability === 0
            ? translate('No seats available in this class. Applications cannot proceed.', 'এই শ্রেণিতে খালি আসন নেই। আবেদন করা যাবে না।')
            : translate('Seat availability is not confirmed. Contact the school office before applying.', 'খালি আসন নিশ্চিত করা হয়নি। আবেদনের আগে স্কুল অফিসে যোগাযোগ করুন।');
    status.dataset.available = String(allowed);
    academicFields.hidden = !rules.needsAcademicCertificate;
    academicFields.disabled = !rules.needsAcademicCertificate;
    if (!rules.needsAcademicCertificate) certificateInputs.forEach(input => {
      input.setCustomValidity('');
      document.getElementById(input.id + 'Error').textContent = '';
    });
    const info = document.getElementById('classRequirements');
    info.textContent = !rules.valid
      ? translate('Choose a class to see its exam and document requirements.', 'পরীক্ষা ও কাগজপত্রের প্রয়োজনীয়তা দেখতে শ্রেণি নির্বাচন করুন।')
      : (rules.needsAcademicCertificate
        ? translate('Transfer certificate and previous academic certificate/report card are required, plus birth certificate, photographs and guardian details.', 'ছাড়পত্র, পূর্ববর্তী শিক্ষাগত সনদ/ফলাফল, জন্ম নিবন্ধন সনদ, ছবি ও অভিভাবকের তথ্য প্রয়োজন।')
        : translate('No previous academic certificate needed for Class 1. Prepare the birth certificate, photographs and guardian details.', '১ম শ্রেণিতে পূর্ববর্তী শিক্ষাগত সনদ প্রয়োজন নেই। জন্ম নিবন্ধন সনদ, ছবি ও অভিভাবকের তথ্য প্রস্তুত রাখুন।'))
        + ' ' + (rules.needsExam
          ? translate('Written admission exam: mandatory for Classes 2–9.', '২য়–৯ম শ্রেণির জন্য লিখিত ভর্তি পরীক্ষা বাধ্যতামূলক।')
          : rules.selection === 'lottery'
            ? translate('Selection: digital lottery; no admission exam for Class 1.', 'নির্বাচন: ডিজিটাল লটারি; ১ম শ্রেণিতে ভর্তি পরীক্ষা নেই।')
            : translate('Class 10: TC/e-TC transfer route, subject to a vacant seat, both schools’ approval and board permission.', '১০ম শ্রেণি: টিসি/ই-টিসির মাধ্যমে বদলি; খালি আসন, উভয় স্কুলের অনুমোদন ও বোর্ডের অনুমতি সাপেক্ষে।'));
  }
  classSelect.addEventListener('change', updateClassRequirements);
  updateClassRequirements();

  const today = new Date();
  const dateString = `${today.getFullYear()}-${String(today.getMonth()+1).padStart(2,'0')}-${String(today.getDate()).padStart(2,'0')}`;
  application.elements.dob.max = dateString;
  application.addEventListener('submit', event => {
    event.preventDefault();
    if (!canStartApplication(classSelect.value)) {
      updateClassRequirements();
      document.getElementById('seatAvailability').scrollIntoView({block:'center'});
      return;
    }
    if (!certificateInputs.map(validateCertificate).every(Boolean)) {
      application.reportValidity();
      return;
    }
    const data = Object.fromEntries([...new FormData(application)].map(([key, value]) => [key, value instanceof File ? value.name : value]));
    const dob = application.elements.dob;
    dob.setCustomValidity(data.dob >= dateString ? translate('Enter a birth date before today.', 'আজকের আগের জন্মতারিখ দিন।') : '');
    if (!application.reportValidity()) return;
    message.hidden = false;
    message.replaceChildren();
    try {
      const records = readDrafts();
      const bytes = crypto.getRandomValues(new Uint32Array(2));
      const reference = `DRAFT-${bytes[0].toString(16)}${bytes[1].toString(16)}`.toUpperCase();
      records[reference] = { ...data, savedAt: new Date().toISOString(), status: 'Local draft — not submitted' };
      localStorage.setItem(storageKey, JSON.stringify(records));
      const heading = document.createElement('h3');
      heading.textContent = translate('Application saved', 'আবেদন সংরক্ষিত');
      const text = document.createElement('p');
      text.textContent = translate(`Reference: ${reference}. Saved on this browser only. No application was sent to the school.`, `রেফারেন্স: ${reference}। শুধু এই ব্রাউজারে সংরক্ষিত। স্কুলে কোনো আবেদন পাঠানো হয়নি।`);
      const download = document.createElement('button');
      download.type = 'button';
      download.textContent = translate('Download application copy', 'আবেদনের কপি ডাউনলোড');
      download.addEventListener('click', () => {
        const receipt = new Blob([`LOCAL APPLICATION DRAFT — NOT SUBMITTED\nReference: ${reference}\n` + Object.entries(data).map(([key,value]) => `${key}: ${value}`).join('\n')], {type:'text/plain;charset=utf-8'});
        const url = URL.createObjectURL(receipt);
        const link = document.createElement('a');
        link.href = url; link.download = `${reference}.txt`; link.click();
        setTimeout(() => URL.revokeObjectURL(url), 1000);
      });
      message.append(heading, text, download);
      message.hidden = false;
    } catch {
      message.textContent = translate('The draft could not be saved. Browser storage may be unavailable. Your form has been kept; try the printable form instead.', 'খসড়া সংরক্ষণ করা যায়নি। ব্রাউজার স্টোরেজ অনুপলব্ধ হতে পারে। আপনার তথ্য ফর্মে আছে; প্রিন্টযোগ্য ফর্ম ব্যবহার করতে পারেন।');
    }
    message.focus();
  });
  application.elements.dob.addEventListener('input', () => application.elements.dob.setCustomValidity(''));
})();
