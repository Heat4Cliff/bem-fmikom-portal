export type HomeAction = {
  label: string;
  href: string;
};

export type LandingHero = {
  eyebrow: string;
  title: string;
  description: string;
  primaryAction: HomeAction;
  secondaryAction: HomeAction;
  visualLabel: string;
  visualCaption: string;
};

export type AboutContent = {
  title: string;
  description: string;
  highlights: { title: string; description: string }[];
};

export type VisionMission = {
  vision: string | null;
  mission: string[];
};

export type OrganizationPreview = {
  periodLabel?: string;
  members: { id: string; role: string; name: string; division?: string }[];
};

export type NewsPreview = {
  id: string;
  category: string;
  date: string;
  title: string;
  excerpt: string;
  visualLabel: string;
  href: string;
};

export type AnnouncementPreview = {
  id: string;
  title: string;
  content: string;
  image?: { src: string; alt: string };
  dateLabel?: string;
  href?: string;
};

export type GalleryPreview = {
  id: string;
  title: string;
  dateLabel?: string;
  division?: string;
  visualLabel: string;
  tone: "navy" | "blue" | "amber" | "slate";
};

export type RegistrationPreview = {
  id: string;
  title: string;
  division?: string;
  openDate?: string;
  closeDate?: string;
  status: string;
};

export type ContactInfo = {
  address?: string;
  email?: string;
  phone?: string;
  socialLinks: { label: string; href: string }[];
};

export type LandingPageContent = {
  demoMode: boolean;
  hero: LandingHero;
  about: AboutContent;
  visionMission: VisionMission;
  statistics: { label: string; value: string }[];
  organization: OrganizationPreview;
  news: NewsPreview[];
  announcements: AnnouncementPreview[];
  gallery: GalleryPreview[];
  registrations: RegistrationPreview[];
  contact: ContactInfo;
};
