# Workflow

## Daily loop (private)
1. Advance world state (creates a snapshot for day N)
2. Generate draft packet:
   - 5 headline/lede candidates
      - 3 full article candidates
      3. Review in console:
         - select / reject with reason tags
            - optional ranking
               - edit selected candidate
               4. Publish article
               5. Log:
                  - review actions
                     - rankings
                        - edit diffs

                        ## Region initialization
                        When a new edition/region is selected:
                        - if region state does not exist, initialize it (baseline stance/modifiers)
                        - persist so every visitor to that region sees the same canon