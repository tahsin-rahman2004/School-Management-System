package main

import "strconv"

type Subject struct{ Name, Bangla string }

func subjectsFor(u User) []Subject {
	class, err := strconv.Atoi(u.Class)
	if err != nil || class < 1 || class > 10 {
		return nil
	}
	subjects := []Subject{{"Bangla", "বাংলা"}, {"English", "ইংরেজি"}, {"Mathematics", "গণিত"}}
	if class <= 2 {
		return subjects
	}
	if class <= 8 {
		subjects = append(subjects, Subject{"Bangladesh and Global Studies", "বাংলাদেশ ও বিশ্বপরিচয়"}, Subject{"Science", "বিজ্ঞান"}, Subject{"Religious Studies", "ধর্ম শিক্ষা"})
		if class >= 6 {
			subjects = append(subjects, Subject{"Information and Communication Technology", "তথ্য ও যোগাযোগ প্রযুক্তি"}, Subject{"Work and Life-oriented Education", "কর্ম ও জীবনমুখী শিক্ষা"}, Subject{"Arts and Culture", "শিল্প ও সংস্কৃতি"}, Subject{"Physical Education and Health", "শারীরিক শিক্ষা ও স্বাস্থ্য"})
		}
		return subjects
	}
	subjects = []Subject{{"Bangla 1st Paper", "বাংলা প্রথম পত্র"}, {"Bangla 2nd Paper", "বাংলা দ্বিতীয় পত্র"}, {"English 1st Paper", "ইংরেজি প্রথম পত্র"}, {"English 2nd Paper", "ইংরেজি দ্বিতীয় পত্র"}, {"General Mathematics", "সাধারণ গণিত"}, {"Information and Communication Technology", "তথ্য ও যোগাযোগ প্রযুক্তি"}, {"Religion and Moral Education", "ধর্ম ও নৈতিক শিক্ষা"}, {"Physical Education and Health", "শারীরিক শিক্ষা ও স্বাস্থ্য"}, {"Bangladesh and Global Studies", "বাংলাদেশ ও বিশ্বপরিচয়"}}
	switch u.Group {
	case "Science":
		subjects = append(subjects, Subject{"Physics", "পদার্থবিজ্ঞান"}, Subject{"Chemistry", "রসায়ন"}, Subject{"Biology", "জীববিজ্ঞান"}, Subject{"Higher Mathematics", "উচ্চতর গণিত"})
	case "Humanities":
		subjects = append(subjects, Subject{"Geography and Environment", "ভূগোল ও পরিবেশ"}, Subject{"Economics", "অর্থনীতি"}, Subject{"Civics and Citizenship", "পৌরনীতি ও নাগরিকতা"}, Subject{"History of Bangladesh and World Civilization", "বাংলাদেশের ইতিহাস ও বিশ্বসভ্যতা"})
	case "Business Studies":
		subjects = append(subjects, Subject{"Accounting", "হিসাববিজ্ঞান"}, Subject{"Finance and Banking", "ফিন্যান্স ও ব্যাংকিং"}, Subject{"Business Entrepreneurship", "ব্যবসায় উদ্যোগ"})
	}
	return subjects
}

func validGroup(group string) bool {
	return group == "Science" || group == "Humanities" || group == "Business Studies"
}
