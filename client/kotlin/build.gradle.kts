plugins {
    kotlin("jvm") version "2.1.0"
}

repositories { mavenCentral() }

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    implementation("com.google.protobuf:protobuf-kotlin:4.28.3")
}

// Generated sources from `buf generate`
sourceSets["main"].java.srcDirs("gen/java", "gen/kotlin")
