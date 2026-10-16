import { AboutSection } from "./components/about-section";
import { AnnouncementSection } from "./components/announcement-section";
import { AspirationSection, ContactSection } from "./components/contact-sections";
import { GallerySection } from "./components/gallery-section";
import { DemoStatistics, HeroSection } from "./components/hero-section";
import { NewsSection } from "./components/news-section";
import { OrganizationSection } from "./components/organization-section";
import { RegistrationSection } from "./components/registration-section";
import type { LandingPageContent } from "./types";
import { FadeContent } from "@/components/react-bits/fade-content";

export function PublicHome({ content }: { content: LandingPageContent }) {
  return (
    <div className="flex flex-col">
      <HeroSection hero={content.hero} demoMode={content.demoMode} />
      <DemoStatistics items={content.statistics} />
      <FadeContent>
        <AboutSection
          about={content.about}
          vision={content.visionMission.vision}
          mission={content.visionMission.mission}
        />
      </FadeContent>
      <FadeContent>
        <OrganizationSection organization={content.organization} demoMode={content.demoMode} />
      </FadeContent>
      <FadeContent>
        <div className="flex flex-col">
          <div className="order-1 lg:order-2">
            <NewsSection news={content.news} demoMode={content.demoMode} />
          </div>
          <div className="order-2 lg:order-1">
            <AnnouncementSection announcements={content.announcements} demoMode={content.demoMode} />
          </div>
        </div>
      </FadeContent>
      <FadeContent>
        <GallerySection gallery={content.gallery} demoMode={content.demoMode} />
      </FadeContent>
      <FadeContent>
        <RegistrationSection registrations={content.registrations} demoMode={content.demoMode} />
      </FadeContent>
      <FadeContent>
        <AspirationSection />
      </FadeContent>
      <FadeContent>
        <ContactSection contact={content.contact} demoMode={content.demoMode} />
      </FadeContent>
    </div>
  );
}
