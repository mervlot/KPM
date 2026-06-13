go run . i io.kvision:kvision-server-javalin:
PS C:\mervlot\mervnote\MervCode\KPM> go run . i io.kvision:kvision-server-javalin:
Get "https://repo1.maven.org/maven2/io/kvision/kvision-common-annotations/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
Get "https://repo1.maven.org/maven2/io/kvision/kvision-common-types/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
Get "https://repo1.maven.org/maven2/io/kvision/kvision-common-remote/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
Get "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-serialization-json/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
Get "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-core/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
Get "https://repo1.maven.org/maven2/org/jetbrains/kotlin/kotlin-stdlib/maven-metadata.xml": dial tcp: lookup repo1.maven.org: no such host
runtime: goroutine stack exceeds 1000000000-byte limit
runtime: sp=0x1bb9b9061330 stack=[0x1bb9b9060000, 0x1bb9d9060000]
fatal error: stack overflow

runtime stack:
runtime.throw({0x7ff7ab459989?, 0x1bb998eba001?})
        C:/Program Files/Go/src/runtime/panic.go:1229 +0x4d fp=0x2235fffcc0 sp=0x2235fffc90 pc=0x7ff7ab11fdad
runtime.newstack()
        C:/Program Files/Go/src/runtime/stack.go:1178 +0x60c fp=0x2235fffdf0 sp=0x2235fffcc0 pc=0x7ff7ab10586c
runtime.morestack()
        C:/Program Files/Go/src/runtime/asm_amd64.s:681 +0x7c fp=0x2235fffdf8 sp=0x2235fffdf0 pc=0x7ff7ab12565c

goroutine 1 gp=0x1bb998eca000 m=3 mp=0x1bb998ed5008 [running]:
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:25 +0x128 fp=0x1bb9b9061340 sp=0x1bb9b9061338 pc=0x7ff7ab32b128
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90613b8 sp=0x1bb9b9061340 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061430 sp=0x1bb9b90613b8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90614a8 sp=0x1bb9b9061430 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061520 sp=0x1bb9b90614a8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061598 sp=0x1bb9b9061520 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061610 sp=0x1bb9b9061598 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061688 sp=0x1bb9b9061610 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061700 sp=0x1bb9b9061688 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061778 sp=0x1bb9b9061700 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90617f0 sp=0x1bb9b9061778 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061868 sp=0x1bb9b90617f0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90618e0 sp=0x1bb9b9061868 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061958 sp=0x1bb9b90618e0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90619d0 sp=0x1bb9b9061958 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061a48 sp=0x1bb9b90619d0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061ac0 sp=0x1bb9b9061a48 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061b38 sp=0x1bb9b9061ac0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061bb0 sp=0x1bb9b9061b38 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061c28 sp=0x1bb9b9061bb0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061ca0 sp=0x1bb9b9061c28 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061d18 sp=0x1bb9b9061ca0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061d90 sp=0x1bb9b9061d18 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061e08 sp=0x1bb9b9061d90 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061e80 sp=0x1bb9b9061e08 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061ef8 sp=0x1bb9b9061e80 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061f70 sp=0x1bb9b9061ef8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9061fe8 sp=0x1bb9b9061f70 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062060 sp=0x1bb9b9061fe8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90620d8 sp=0x1bb9b9062060 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062150 sp=0x1bb9b90620d8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90621c8 sp=0x1bb9b9062150 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062240 sp=0x1bb9b90621c8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90622b8 sp=0x1bb9b9062240 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062330 sp=0x1bb9b90622b8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90623a8 sp=0x1bb9b9062330 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062420 sp=0x1bb9b90623a8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062498 sp=0x1bb9b9062420 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062510 sp=0x1bb9b9062498 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062588 sp=0x1bb9b9062510 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062600 sp=0x1bb9b9062588 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062678 sp=0x1bb9b9062600 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90626f0 sp=0x1bb9b9062678 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062768 sp=0x1bb9b90626f0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90627e0 sp=0x1bb9b9062768 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062858 sp=0x1bb9b90627e0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90628d0 sp=0x1bb9b9062858 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062948 sp=0x1bb9b90628d0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b90629c0 sp=0x1bb9b9062948 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9b9062a38 sp=0x1bb9b90629c0 pc=0x7ff7ab32b065
...4473780 frames elided...
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e710 sp=0x1bb9d905e698 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e788 sp=0x1bb9d905e710 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e800 sp=0x1bb9d905e788 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e878 sp=0x1bb9d905e800 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e8f0 sp=0x1bb9d905e878 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e968 sp=0x1bb9d905e8f0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905e9e0 sp=0x1bb9d905e968 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ea58 sp=0x1bb9d905e9e0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ead0 sp=0x1bb9d905ea58 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905eb48 sp=0x1bb9d905ead0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ebc0 sp=0x1bb9d905eb48 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ec38 sp=0x1bb9d905ebc0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ecb0 sp=0x1bb9d905ec38 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ed28 sp=0x1bb9d905ecb0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905eda0 sp=0x1bb9d905ed28 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ee18 sp=0x1bb9d905eda0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ee90 sp=0x1bb9d905ee18 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ef08 sp=0x1bb9d905ee90 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905ef80 sp=0x1bb9d905ef08 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905eff8 sp=0x1bb9d905ef80 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f070 sp=0x1bb9d905eff8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f0e8 sp=0x1bb9d905f070 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f160 sp=0x1bb9d905f0e8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f1d8 sp=0x1bb9d905f160 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f250 sp=0x1bb9d905f1d8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f2c8 sp=0x1bb9d905f250 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f340 sp=0x1bb9d905f2c8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f3b8 sp=0x1bb9d905f340 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f430 sp=0x1bb9d905f3b8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f4a8 sp=0x1bb9d905f430 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f520 sp=0x1bb9d905f4a8 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f598 sp=0x1bb9d905f520 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f610 sp=0x1bb9d905f598 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f688 sp=0x1bb9d905f610 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f700 sp=0x1bb9d905f688 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f778 sp=0x1bb9d905f700 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f7f0 sp=0x1bb9d905f778 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f868 sp=0x1bb9d905f7f0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f8e0 sp=0x1bb9d905f868 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f958 sp=0x1bb9d905f8e0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905f9d0 sp=0x1bb9d905f958 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905fa48 sp=0x1bb9d905f9d0 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905fac0 sp=0x1bb9d905fa48 pc=0x7ff7ab32b065
kpm/types.Mavenurl.BuildPath({{0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e8ed30, 0x5}}, {0x0, 0x0})
        C:/mervlot/mervnote/MervCode/KPM/types/maven.go:26 +0x65 fp=0x1bb9d905fb38 sp=0x1bb9d905fac0 pc=0x7ff7ab32b065
kpm/package.DownloadMavenInternal({0x1bb998e92390, 0xa}, {0x1bb998e9239b, 0x16}, {0x1bb998e9239b, 0x0}, 0x88?, 0x7ff7ab741fa0, 0x0?, 0x1)
        C:/mervlot/mervnote/MervCode/KPM/package/downloadPackage.go:85 +0x80d fp=0x1bb9d905fcf0 sp=0x1bb9d905fb38 pc=0x7ff7ab349c4d
kpm/package.DownloadMaven(...)
        C:/mervlot/mervnote/MervCode/KPM/package/downloadPackage.go:13
kpm/package.Main(0x0, {0x1bb998ee00a0, 0x1, 0x2000000000080?})
        C:/mervlot/mervnote/MervCode/KPM/package/main.go:122 +0x755 fp=0x1bb9d905fe70 sp=0x1bb9d905fcf0 pc=0x7ff7ab34a8b5
main.main()
        C:/mervlot/mervnote/MervCode/KPM/main.go:54 +0x3d9 fp=0x1bb9d905ff48 sp=0x1bb9d905fe70 pc=0x7ff7ab3529b9
runtime.main()
        C:/Program Files/Go/src/runtime/proc.go:290 +0x2c7 fp=0x1bb9d905ffe0 sp=0x1bb9d905ff48 pc=0x7ff7ab0eda07
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb9d905ffe8 sp=0x1bb9d905ffe0 pc=0x7ff7ab1271c1

goroutine 2 gp=0x1bb998eca780 m=nil [force gc (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998ecdfa8 sp=0x1bb998ecdf88 pc=0x7ff7ab11fece
runtime.goparkunlock(...)
        C:/Program Files/Go/src/runtime/proc.go:468
runtime.forcegchelper()
        C:/Program Files/Go/src/runtime/proc.go:375 +0xb8 fp=0x1bb998ecdfe0 sp=0x1bb998ecdfa8 pc=0x7ff7ab0edcf8
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998ecdfe8 sp=0x1bb998ecdfe0 pc=0x7ff7ab1271c1
created by runtime.init.7 in goroutine 1
        C:/Program Files/Go/src/runtime/proc.go:363 +0x1a

goroutine 3 gp=0x1bb998ecab40 m=nil [GC sweep wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998ecff88 sp=0x1bb998ecff68 pc=0x7ff7ab11fece
runtime.goparkunlock(...)
        C:/Program Files/Go/src/runtime/proc.go:468
runtime.bgsweep(0x1bb998e9c080)
        C:/Program Files/Go/src/runtime/mgcsweep.go:279 +0x94 fp=0x1bb998ecffc8 sp=0x1bb998ecff88 pc=0x7ff7ab0d63d4
runtime.gcenable.gowrap1()
        C:/Program Files/Go/src/runtime/mgc.go:214 +0x17 fp=0x1bb998ecffe0 sp=0x1bb998ecffc8 pc=0x7ff7ab0c79d7
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998ecffe8 sp=0x1bb998ecffe0 pc=0x7ff7ab1271c1
created by runtime.gcenable in goroutine 1
        C:/Program Files/Go/src/runtime/mgc.go:214 +0x66

goroutine 4 gp=0x1bb998ecad20 m=nil [GC scavenge wait]:
runtime.gopark(0x1bb998e9c080?, 0x7ff7ab4784e0?, 0x1?, 0x0?, 0x1bb998ecad20?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998eddf78 sp=0x1bb998eddf58 pc=0x7ff7ab11fece
runtime.goparkunlock(...)
        C:/Program Files/Go/src/runtime/proc.go:468
runtime.(*scavengerState).park(0x7ff7ab752760)
        C:/Program Files/Go/src/runtime/mgcscavenge.go:425 +0x49 fp=0x1bb998eddfa8 sp=0x1bb998eddf78 pc=0x7ff7ab0d3f09
runtime.bgscavenge(0x1bb998e9c080)
        C:/Program Files/Go/src/runtime/mgcscavenge.go:653 +0x3c fp=0x1bb998eddfc8 sp=0x1bb998eddfa8 pc=0x7ff7ab0d447c
runtime.gcenable.gowrap2()
        C:/Program Files/Go/src/runtime/mgc.go:215 +0x17 fp=0x1bb998eddfe0 sp=0x1bb998eddfc8 pc=0x7ff7ab0c7997
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998eddfe8 sp=0x1bb998eddfe0 pc=0x7ff7ab1271c1
created by runtime.gcenable in goroutine 1
        C:/Program Files/Go/src/runtime/mgc.go:215 +0xa5

goroutine 5 gp=0x1bb998ecb0e0 m=nil [GOMAXPROCS updater (idle)]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998edff88 sp=0x1bb998edff68 pc=0x7ff7ab11fece
runtime.goparkunlock(...)
        C:/Program Files/Go/src/runtime/proc.go:468
runtime.updateMaxProcsGoroutine()
        C:/Program Files/Go/src/runtime/proc.go:7095 +0xe7 fp=0x1bb998edffe0 sp=0x1bb998edff88 pc=0x7ff7ab0fbc67
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998edffe8 sp=0x1bb998edffe0 pc=0x7ff7ab1271c1
created by runtime.defaultGOMAXPROCSUpdateEnable in goroutine 1
        C:/Program Files/Go/src/runtime/proc.go:7083 +0x37

goroutine 6 gp=0x1bb998ecb2c0 m=nil [finalizer wait]:
runtime.gopark(0x0?, 0x0?, 0x0?, 0x0?, 0x0?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998ed9e20 sp=0x1bb998ed9e00 pc=0x7ff7ab11fece
runtime.runFinalizers()
        C:/Program Files/Go/src/runtime/mfinal.go:210 +0x107 fp=0x1bb998ed9fe0 sp=0x1bb998ed9e20 pc=0x7ff7ab0c6967
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998ed9fe8 sp=0x1bb998ed9fe0 pc=0x7ff7ab1271c1
created by runtime.createfing in goroutine 1
        C:/Program Files/Go/src/runtime/mfinal.go:172 +0x3d

goroutine 7 gp=0x1bb998ecb4a0 m=nil [cleanup wait]:
runtime.gopark(0x101000000000000?, 0x5d?, 0x3?, 0x0?, 0x7ff7ab6f35a0?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998ed1f68 sp=0x1bb998ed1f48 pc=0x7ff7ab11fece
runtime.goparkunlock(...)
        C:/Program Files/Go/src/runtime/proc.go:468
runtime.(*cleanupQueue).dequeue(0x7ff7ab752980)
        C:/Program Files/Go/src/runtime/mcleanup.go:522 +0xd4 fp=0x1bb998ed1fa0 sp=0x1bb998ed1f68 pc=0x7ff7ab0c3954
runtime.runCleanups()
        C:/Program Files/Go/src/runtime/mcleanup.go:718 +0x45 fp=0x1bb998ed1fe0 sp=0x1bb998ed1fa0 pc=0x7ff7ab0c3fc5
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998ed1fe8 sp=0x1bb998ed1fe0 pc=0x7ff7ab1271c1
created by runtime.(*cleanupQueue).createGs in goroutine 1
        C:/Program Files/Go/src/runtime/mcleanup.go:672 +0xa5

goroutine 8 gp=0x1bb998ecb680 m=nil [select]:
runtime.gopark(0x1bb998edbf90?, 0x2?, 0x70?, 0x0?, 0x1bb998edbf84?)
        C:/Program Files/Go/src/runtime/proc.go:462 +0xce fp=0x1bb998edbe20 sp=0x1bb998edbe00 pc=0x7ff7ab11fece
runtime.selectgo(0x1bb998edbf90, 0x1bb998edbf80, 0x0?, 0x0, 0x0?, 0x1)
        C:/Program Files/Go/src/runtime/select.go:351 +0xaa5 fp=0x1bb998edbf50 sp=0x1bb998edbe20 pc=0x7ff7ab0fff05
github.com/patrickmn/go-cache.(*janitor).Run(0x1bb998eac6d0, 0x1bb998ee1240)
        C:/Users/MRS WURAOLA/go/pkg/mod/github.com/patrickmn/go-cache@v2.1.0+incompatible/cache.go:1079 +0x7d fp=0x1bb998edbfc0 sp=0x1bb998edbf50 pc=0x7ff7ab32609d
github.com/patrickmn/go-cache.runJanitor.gowrap1()
        C:/Users/MRS WURAOLA/go/pkg/mod/github.com/patrickmn/go-cache@v2.1.0+incompatible/cache.go:1099 +0x1b fp=0x1bb998edbfe0 sp=0x1bb998edbfc0 pc=0x7ff7ab32623b
runtime.goexit({})
        C:/Program Files/Go/src/runtime/asm_amd64.s:1771 +0x1 fp=0x1bb998edbfe8 sp=0x1bb998edbfe0 pc=0x7ff7ab1271c1
created by github.com/patrickmn/go-cache.runJanitor in goroutine 1
        C:/Users/MRS WURAOLA/go/pkg/mod/github.com/patrickmn/go-cache@v2.1.0+incompatible/cache.go:1099 +0xdb
exit status 2
