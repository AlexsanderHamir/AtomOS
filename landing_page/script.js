// Smooth scrolling for navigation links
document.addEventListener("DOMContentLoaded", function () {
  // Mobile menu toggle
  const hamburger = document.querySelector(".hamburger");
  const navMenu = document.querySelector(".nav-menu");

  if (hamburger && navMenu) {
    hamburger.addEventListener("click", function () {
      navMenu.classList.toggle("active");
      hamburger.classList.toggle("active");

      // Prevent body scroll when menu is open
      if (navMenu.classList.contains("active")) {
        document.body.style.overflow = "hidden";
      } else {
        document.body.style.overflow = "auto";
      }
    });

    // Close menu when clicking on nav links
    const navLinks = document.querySelectorAll(".nav-menu .nav-link");
    navLinks.forEach((link) => {
      link.addEventListener("click", () => {
        navMenu.classList.remove("active");
        hamburger.classList.remove("active");
        document.body.style.overflow = "auto";
      });
    });

    // Close menu when clicking outside
    document.addEventListener("click", function (event) {
      if (
        !hamburger.contains(event.target) &&
        !navMenu.contains(event.target)
      ) {
        navMenu.classList.remove("active");
        hamburger.classList.remove("active");
        document.body.style.overflow = "auto";
      }
    });
  }

  // Smooth scrolling for anchor links
  const navLinks = document.querySelectorAll('a[href^="#"]');
  navLinks.forEach((link) => {
    link.addEventListener("click", function (e) {
      e.preventDefault();
      const targetId = this.getAttribute("href");
      const targetSection = document.querySelector(targetId);

      if (targetSection) {
        const offsetTop = targetSection.offsetTop - 70; // Account for fixed navbar
        window.scrollTo({
          top: offsetTop,
          behavior: "smooth",
        });
      }
    });
  });

  // Navbar background on scroll
  const navbar = document.querySelector(".navbar");
  window.addEventListener("scroll", function () {
    if (window.scrollY > 50) {
      navbar.style.background = "rgba(255, 255, 255, 0.98)";
      navbar.style.boxShadow = "0 2px 20px rgba(0, 0, 0, 0.1)";
    } else {
      navbar.style.background = "rgba(255, 255, 255, 0.95)";
      navbar.style.boxShadow = "none";
    }
  });

  // Intersection Observer for animations
  const observerOptions = {
    threshold: 0.1,
    rootMargin: "0px 0px -50px 0px",
  };

  const observer = new IntersectionObserver(function (entries) {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("animate-fade-in-up");
      }
    });
  }, observerOptions);

  // Observe elements for animation
  const animateElements = document.querySelectorAll(
    ".feature-card, .example-card, .step"
  );
  animateElements.forEach((el) => {
    observer.observe(el);
  });

  // Workflow diagram animation
  const workflowBlocks = document.querySelectorAll(".block");
  workflowBlocks.forEach((block, index) => {
    block.style.animationDelay = `${index * 0.2}s`;
    block.classList.add("animate-fade-in-up");
  });

  // Counter animation for stats
  const stats = document.querySelectorAll(".stat-number");
  const statsObserver = new IntersectionObserver(
    function (entries) {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          animateCounter(entry.target);
          statsObserver.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.5 }
  );

  stats.forEach((stat) => {
    statsObserver.observe(stat);
  });

  // Counter animation function
  function animateCounter(element) {
    const target = element.textContent;
    const isPercentage = target.includes("%");
    const isMultiplier = target.includes("x");
    const numericValue = parseFloat(target.replace(/[^\d.]/g, ""));

    let current = 0;
    const increment = numericValue / 50;
    const timer = setInterval(() => {
      current += increment;
      if (current >= numericValue) {
        current = numericValue;
        clearInterval(timer);
      }

      let displayValue = Math.floor(current);
      if (isPercentage) {
        element.textContent = displayValue + "%";
      } else if (isMultiplier) {
        element.textContent = displayValue + "x";
      } else if (target.includes("+")) {
        element.textContent = displayValue + "+";
      } else {
        element.textContent = displayValue;
      }
    }, 30);
  }

  // Parallax effect for hero section
  const hero = document.querySelector(".hero");
  if (hero) {
    window.addEventListener("scroll", function () {
      const scrolled = window.pageYOffset;
      const rate = scrolled * -0.5;
      hero.style.transform = `translateY(${rate}px)`;
    });
  }

  // Typing effect removed - text appears immediately

  // Button hover effects
  const buttons = document.querySelectorAll("button");
  buttons.forEach((button) => {
    button.addEventListener("mouseenter", function () {
      this.style.transform = "translateY(-2px)";
    });

    button.addEventListener("mouseleave", function () {
      this.style.transform = "translateY(0)";
    });
  });

  // Card hover effects
  const cards = document.querySelectorAll(".feature-card, .example-card");
  cards.forEach((card) => {
    card.addEventListener("mouseenter", function () {
      this.style.transform = "translateY(-5px) scale(1.02)";
    });

    card.addEventListener("mouseleave", function () {
      this.style.transform = "translateY(0) scale(1)";
    });
  });

  // Code snippet copy functionality
  const codeSnippets = document.querySelectorAll(".code-snippet");
  codeSnippets.forEach((snippet) => {
    snippet.addEventListener("click", function () {
      const code = this.querySelector("code").textContent;
      navigator.clipboard.writeText(code).then(() => {
        // Show feedback
        const originalBg = this.style.backgroundColor;
        this.style.backgroundColor = "#10b981";
        setTimeout(() => {
          this.style.backgroundColor = originalBg;
        }, 200);
      });
    });

    // Add copy indicator
    snippet.style.cursor = "pointer";
    snippet.title = "Click to copy";
  });

  // Lazy loading for images (if any are added later)
  const images = document.querySelectorAll("img[data-src]");
  const imageObserver = new IntersectionObserver((entries, observer) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const img = entry.target;
        img.src = img.dataset.src;
        img.classList.remove("lazy");
        imageObserver.unobserve(img);
      }
    });
  });

  images.forEach((img) => imageObserver.observe(img));

  // Waiting list form handling
  const waitingListForm = document.getElementById("waitingListForm");
  const formSuccess = document.getElementById("formSuccess");

  if (waitingListForm) {
    waitingListForm.addEventListener("submit", function (e) {
      e.preventDefault();

      const emailInput = document.getElementById("email");
      const submitButton = this.querySelector(".submit-button");
      const email = emailInput.value.trim();

      if (!email) {
        return;
      }

      // Show loading state
      submitButton.innerHTML =
        '<i class="fas fa-spinner fa-spin"></i> Joining...';
      submitButton.disabled = true;

      // Simulate API call (replace with actual endpoint)
      setTimeout(() => {
        // Hide form and show success message
        this.style.display = "none";
        formSuccess.style.display = "flex";

        // Reset form for potential future use
        emailInput.value = "";
        submitButton.innerHTML =
          '<i class="fas fa-paper-plane"></i> Join Waitlist';
        submitButton.disabled = false;

        // TODO: Replace with API call
        // Log the email (replace with actual API call)
        console.log("Email added to waitlist:", email);

        // Optional: Track analytics event
        if (typeof gtag !== "undefined") {
          gtag("event", "waitlist_signup", {
            event_category: "engagement",
            event_label: "cloud_service_waitlist",
          });
        }
      }, 1500);
    });
  }

  // Scroll to top functionality
  const scrollToTopBtn = document.createElement("button");
  scrollToTopBtn.innerHTML = '<i class="fas fa-arrow-up"></i>';
  scrollToTopBtn.className = "scroll-to-top";
  scrollToTopBtn.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        width: 50px;
        height: 50px;
        border-radius: 50%;
        background: var(--primary-color);
        color: white;
        border: none;
        cursor: pointer;
        opacity: 0;
        transition: all 0.3s ease;
        z-index: 1000;
        box-shadow: var(--shadow-lg);
    `;

  document.body.appendChild(scrollToTopBtn);

  window.addEventListener("scroll", function () {
    if (window.pageYOffset > 300) {
      scrollToTopBtn.style.opacity = "1";
    } else {
      scrollToTopBtn.style.opacity = "0";
    }
  });

  scrollToTopBtn.addEventListener("click", function () {
    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  });

  // Performance optimization: Debounce scroll events
  function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  // Apply debouncing to scroll events
  const debouncedScrollHandler = debounce(function () {
    // Scroll-based animations and effects
    const scrolled = window.pageYOffset;
    const parallaxElements = document.querySelectorAll(".hero");

    parallaxElements.forEach((element) => {
      const rate = scrolled * -0.5;
      element.style.transform = `translateY(${rate}px)`;
    });
  }, 10);

  window.addEventListener("scroll", debouncedScrollHandler);

  // Add loading states for buttons
  const ctaButtons = document.querySelectorAll(
    ".primary-button, .secondary-button"
  );
  ctaButtons.forEach((button) => {
    button.addEventListener("click", function () {
      if (!this.classList.contains("loading")) {
        this.classList.add("loading");
        this.style.pointerEvents = "none";

        // Simulate loading (replace with actual action)
        setTimeout(() => {
          this.classList.remove("loading");
          this.style.pointerEvents = "auto";
        }, 2000);
      }
    });
  });

  // Keyboard navigation support
  document.addEventListener("keydown", function (e) {
    if (e.key === "Tab") {
      document.body.classList.add("keyboard-navigation");
    }
  });

  document.addEventListener("mousedown", function () {
    document.body.classList.remove("keyboard-navigation");
  });

  // Add focus styles for keyboard navigation
  const style = document.createElement("style");
  style.textContent = `
        .keyboard-navigation *:focus {
            outline: 2px solid var(--primary-color) !important;
            outline-offset: 2px !important;
        }
    `;
  document.head.appendChild(style);

  // Initialize tooltips (if needed)
  const tooltipElements = document.querySelectorAll("[data-tooltip]");
  tooltipElements.forEach((element) => {
    element.addEventListener("mouseenter", function () {
      const tooltip = document.createElement("div");
      tooltip.className = "tooltip";
      tooltip.textContent = this.dataset.tooltip;
      tooltip.style.cssText = `
                position: absolute;
                background: var(--bg-dark);
                color: white;
                padding: 8px 12px;
                border-radius: 4px;
                font-size: 14px;
                z-index: 1000;
                pointer-events: none;
                opacity: 0;
                transition: opacity 0.3s ease;
            `;

      document.body.appendChild(tooltip);

      const rect = this.getBoundingClientRect();
      tooltip.style.left =
        rect.left + rect.width / 2 - tooltip.offsetWidth / 2 + "px";
      tooltip.style.top = rect.top - tooltip.offsetHeight - 8 + "px";

      setTimeout(() => {
        tooltip.style.opacity = "1";
      }, 10);

      this.addEventListener(
        "mouseleave",
        function () {
          tooltip.remove();
        },
        { once: true }
      );
    });
  });

  console.log("AtomOS landing page initialized successfully! 🚀");
});
