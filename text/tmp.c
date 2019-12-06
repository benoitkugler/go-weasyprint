type = gUnicharType(wc);
jamo = JAMOTYPE(breakType);

/* Determine wheter this forms a Hangul syllable with prev. */
if (jamo == NOJAMO)
    makesHangulSyllable = FALSE;
else
{
    JamoType prevEnd = HangulJamoProps[prevJamo].end;
    JamoType thisStart = HangulJamoProps[jamo].start;

    /* See comments before ISJAMO */
    makesHangulSyllable = (prevEnd == thisStart) || (prevEnd + 1 == thisStart);
}